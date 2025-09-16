package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/cloud"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
)

type ColumnMeta struct {
	SchemaName        string         `json:"schema_name"`
	TableName         string         `json:"table_name"`
	TableDescription  sql.NullString `json:"table_description"`
	ColumnName        string         `json:"column_name"`
	ColumnType        string         `json:"column_type"`
	NotNull           bool           `json:"not_null"`
	DefaultValue      sql.NullString `json:"default_value"`
	ColumnDescription sql.NullString `json:"column_description"`
	ColumnRole        string         `json:"column_role"` // primary_key | foreign_key | none
	Ref               sql.NullString `json:"ref"`         // <schema>.<table>.<column> or NULL
}

const SYSTEM_INSTRUCTIONS_NO_EDGES = `
You are a diagram-generating assistant (React Flow). Given a JSON array of
table descriptors, output JavaScript that declares exactly two constants:

const initialNodes = [/* …nodes… */];
const initialEdges = [];

INPUT FORMAT
The user supplies a JSON array where each element is:
{
  "label": "<TableName>",
  "schema": [
    {
      "col_role": "primary_key" | "foreign_key" | "none",
      "title": "<ColumnName>",
      "type": "<ColumnType>",
      "ref": "<schema.table.column>" | null
    },
    …
  ]
}

OUTPUT RULES
1. Output only the two constant declarations — no surrounding prose.
2. initialEdges must be an empty array ([]).
3. For every table create one node object:
   {
     id: "<uuid-v4>",
     type: "database_schema",
     position: { x: <int>, y: <int> },    // units are pixels
     style: { zIndex: <unique-int> },     // start at 1000, +1000 per node
     data: {
       type: "database_schema",
       label: "<TableName>",
       schema: [ … ]                      // copy schema verbatim, omit "ref"
     },
     selected: false
   }

GEOMETRY
• Node width defaults to 240 px; if the table name or any column type is
  multi-word it may grow up to 400 px (rough estimate is fine).
• Node height = 78 px + (33 px * schema.length).
  You usually place nodes correctly x-wise but on top of each other by y coordinate.
  Always calculate node total height to prevent this overlap.
• Maintain **at least 200 px** clear space between every pair of node
  *bounding boxes* horizontally **and** vertically — nodes must never
  overlap or touch.

LAYOUT & GROUPING
• Use the ref fields and table names to infer relationships.
  - Place each direct child (FK to a parent's PK) **directly below its
    parent** at the next available row (same column block).
  - If multiple tables share the same parent and have no other relations,
    align them in the **same horizontal row** (siblings) with ≥200 px
    gaps between their bounding boxes.
  - One-to-many and many-to-many bridge tables should be kept in the same
    column block or an adjacent one.
• Cluster related tables tightly; push unrelated clusters farther apart
  so the groupings are visually distinct.
• It's a relational database schema. The most important thing is to make
  relationships clear and you need to be smart about it.
  It's allowed to even break some rules to achieve the best layout.

MISC
• No trailing commas; output must be valid JavaScript.
• Do not include comments inside the emitted JS.

If the input array is empty, output exactly:

const initialNodes = [];
const initialEdges = [];
`

// System prompt used by the diagram-building assistant.
const SYSTEM_INSTRUCTIONS = `
You are a diagram-generating assistant (React Flow). Given a JSON array of
table descriptors, output JavaScript that declares exactly two constants:

const initialNodes = [/* …nodes… */];
const initialEdges = [/* …edges… */];

INPUT FORMAT
The user supplies a JSON array where each element is:
{
  "label": "<TableName>",
  "schema": [
    {
      "col_role": "primary_key" | "foreign_key" | "none",
      "title": "<ColumnName>",
      "type": "<ColumnType>",
      "ref": "<schema.table.column>" | null
    },
    …
  ]
}

OUTPUT RULES
1. Output only the two constant declarations — no surrounding prose.
2. For every table create one node object:
   {
     id: "<uuid-v4>",
     type: "database_schema",
     position: { x: <int>, y: <int> },    // units are pixels
     style: { zIndex: <unique-int> },     // start at 1000, +1000 per node
     data: {
       type: "database_schema",
       label: "<TableName>",
       schema: [ … ]                      // copy schema verbatim, omit "ref"
     },
     selected: false
   }

GEOMETRY
• Node width defaults to 240 px; if the table name or any column type is
  multi-word it may grow up to 400 px (rough estimate is fine).
• Node height = 78 px + (33 px * schema.length).
  You usually place nodes correctly x-wise but on top of each other by y coordinate.
  Always calculate node total height to prevent this overlap.
• Maintain **at least 200 px** clear space between every pair of node
  *bounding boxes* horizontally **and** vertically — nodes must never
  overlap or touch.

LAYOUT & GROUPING
• Use the ref fields and table names to infer relationships.
  - Place each direct child (FK to a parent's PK) **directly below its
    parent** at the next available row (same column block).
  - If multiple tables share the same parent and have no other relations,
    align them in the **same horizontal row** (siblings) with ≥200 px
    gaps between their bounding boxes.
  - One-to-many and many-to-many bridge tables should be kept in the same
    column block or an adjacent one.
• Cluster related tables tightly; push unrelated clusters farther apart
  so the groupings are visually distinct.
• It's a relational database schema. The most important thing is to make
  relationships clear and you need to be smart about it.
  It's allowed to even break some rules to achieve the best layout.

MISC
• No trailing commas; output must be valid JavaScript.
• Do not include comments inside the emitted JS.

EDGES
• For each **direct** foreign-key relationship (child → parent) add exactly
  one edge object and include it in the initialEdges array.
  Skip edges that would be transitively redundant
  (e.g. if C→B and B→A, do NOT also add C→A).
• Edge object format:

  {
    style: {
      strokeWidth: 2,
      strokeDasharray: "5,5",
      stroke: "#b1b1b7"
    },
    markerEnd: { type: "arrowclosed", color: "#b1b1b7" },
    source: "<child-node-uuid>",
    sourceHandle: "<child-node-schema-item-title>",
    target: "<parent-node-uuid>",
    targetHandle: "<parent-node-schema-item-title>",
    id: "<uuid-v4>",
    type: "default",
    selected: false,
    animated: true,
    data: { algorithm: "Bezier Catmull-Rom", points: [] }
  }

If the input array is empty, output exactly:

const initialNodes = [];
const initialEdges = [];
`

// getSchemaMetadata runs the catalog query for one schema and returns per-column metadata.
func getSchemaMetadata(db *sql.DB, schemaNames []string) ([]ColumnMeta, error) {
	var args []interface{}
	var whereClause string

	if len(schemaNames) > 0 {
		placeholders := make([]string, len(schemaNames))
		for i, name := range schemaNames {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, name)
		}
		whereClause = fmt.Sprintf("AND n.nspname IN (%s)", strings.Join(placeholders, ", "))
	} else {
		whereClause = "" // no filtering, match all schemas
	}

	query := fmt.Sprintf(`
		WITH column_constraints AS (
			SELECT
				conrelid,
				unnest(conkey) AS attnum,
				contype,
				confrelid,
				unnest(COALESCE(confkey, ARRAY[NULL::int])) AS confattnum
			FROM pg_constraint
			WHERE contype IN ('p', 'f')
		), fk_targets AS (
			SELECT
				c.oid   AS relid,
				a.attnum,
				n.nspname AS ref_schema,
				c.relname AS ref_table,
				a.attname AS ref_column
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			JOIN pg_attribute a ON a.attrelid = c.oid
			WHERE a.attnum > 0 AND NOT a.attisdropped
		)
		SELECT
			n.nspname,
			c.relname,
			obj_description(c.oid, 'pg_class')                      AS table_description,
			a.attname,
			format_type(a.atttypid, a.atttypmod)                    AS column_type,
			a.attnotnull,
			pg_get_expr(ad.adbin, ad.adrelid)                       AS default_value,
			col_description(a.attrelid, a.attnum)                   AS column_description,
			CASE cc.contype WHEN 'p' THEN 'primary_key'
							WHEN 'f' THEN 'foreign_key'
							ELSE 'none' END                         AS column_role,
			CASE WHEN cc.contype = 'f'
				THEN ft.ref_schema || '.' || ft.ref_table || '.' || ft.ref_column
				ELSE NULL END                                      AS ref
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid
		LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
		LEFT JOIN column_constraints cc
			ON cc.conrelid = c.oid AND cc.attnum = a.attnum
		LEFT JOIN fk_targets ft
			ON cc.contype = 'f' AND cc.confrelid = ft.relid AND cc.confattnum = ft.attnum
		WHERE c.relkind = 'r'
		AND a.attnum > 0
		AND NOT a.attisdropped
		%s
		ORDER BY c.relname, a.attnum;
	`, whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying metadata: %w", err)
	}
	defer rows.Close()

	var out []ColumnMeta
	for rows.Next() {
		var m ColumnMeta
		if err := rows.Scan(
			&m.SchemaName,
			&m.TableName,
			&m.TableDescription,
			&m.ColumnName,
			&m.ColumnType,
			&m.NotNull,
			&m.DefaultValue,
			&m.ColumnDescription,
			&m.ColumnRole,
			&m.Ref,
		); err != nil {
			return nil, fmt.Errorf("scanning metadata: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating metadata rows: %w", err)
	}
	return out, nil
}

func buildSchemasJSON(cols []ColumnMeta) ([]byte, error) {
	type columnDef struct {
		ColRole string `json:"col_role"`
		Title   string `json:"title"`
		Type    string `json:"type"`
		Ref     string `json:"ref,omitempty"` // present if foreign key
	}

	type tableSchema struct {
		Label  string      `json:"label"`
		Schema []columnDef `json:"schema"`
	}

	group := make(map[string][]columnDef)
	for _, c := range cols {
		def := columnDef{
			ColRole: c.ColumnRole,
			Title:   c.ColumnName,
			Type:    c.ColumnType,
			Ref:     c.Ref.String,
		}
		group[c.TableName] = append(group[c.TableName], def)
	}

	var out []tableSchema
	for tbl, defs := range group {
		out = append(out, tableSchema{Label: tbl, Schema: defs})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })

	return json.Marshal(out)
}

func generateDiagram(tenantID string, userID string, schemaJSON []byte) (*string, error) {
	client, _, err := GetOpenAiClient(tenantID)
	if err != nil {
		return nil, err
	}

	var params = &responses.ResponseNewParams{
		Model:        "o3",
		User:         openai.String(userID),
		Instructions: openai.String(SYSTEM_INSTRUCTIONS),
		// Temperature:     openai.Float(0.5),
		// TopP:            openai.Float(1),
		MaxOutputTokens: openai.Int(16000),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(string(schemaJSON)),
		},
	}

	response, err := client.Responses.New(
		context.Background(),
		*params,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat response: %v", err)
	}

	defer func() {
		log.Printf("Token usage: prompt=%d, completion=%d", response.Usage.InputTokens, response.Usage.OutputTokens)

		err := newTokenUsageResource(&models.TenantTokenUsageRequest{
			TenantID:         tenantID,
			UserID:           userID,
			AiVendor:         "openai",
			AiModel:          response.Model,
			Tools:            map[string]interface{}{},
			PromptTokens:     int32(response.Usage.InputTokens),
			CompletionTokens: int32(response.Usage.OutputTokens),
		})
		if err == nil {
			log.Printf("Token usage stored successfully")
		}

	}()

	// TODO: Validate output text for valid JavaScript
	var outputText = response.OutputText()

	return &outputText, nil
}

func GenerateDatabaseDesignHandler(c *gin.Context) {
	startTotal := time.Now()

	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	type Payload struct {
		CredentialID string   `json:"credential_id"`
		SchemaNames  []string `json:"schema_names"`
	}
	var payload Payload

	startParse := time.Now()
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	log.Printf("Parsed payload in %v", time.Since(startParse))

	startConnect := time.Now()
	db, err := cloud.ConnectToPostgres(tenantID, payload.CredentialID)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to PostgreSQL"})
		return
	}
	defer db.Close()
	log.Printf("Connected to PostgreSQL in %v", time.Since(startConnect))

	startMetadata := time.Now()
	schemaData, err := getSchemaMetadata(db, payload.SchemaNames)
	if err != nil {
		log.Printf("Failed to get schema metadata: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}
	log.Printf("Fetched schema metadata in %v", time.Since(startMetadata))

	if len(schemaData) == 0 {
		log.Printf("No schema metadata found for tenant %s. Schemas: %v", tenantID, payload.SchemaNames)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No tables found. Please verify that your database contains tables or check the provided schema name(s).",
		})
		return
	}

	startBuild := time.Now()
	schemaJSON, err := buildSchemasJSON(schemaData)
	if err != nil {
		log.Printf("Failed to build schema JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}
	log.Printf("Built schema JSON in %v", time.Since(startBuild))

	startGen := time.Now()
	diagramJS, err := generateDiagram(tenantID, userID, schemaJSON)
	if err != nil {
		log.Printf("Failed to generate diagram: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}
	log.Printf("Generated diagram in %v", time.Since(startGen))
	log.Printf("Total time: %v", time.Since(startTotal))

	c.JSON(http.StatusOK, gin.H{
		"data":    *diagramJS,
		"message": "Diagram generated successfully",
	})
}
