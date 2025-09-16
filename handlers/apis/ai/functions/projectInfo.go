package aiFunctions

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProjectData struct {
	Title        string
	Status       string
	Category     string
	Description  sql.NullString
	Complexity   sql.NullString
	Requirements []RequirementData
}

type RequirementData struct {
	Title    string
	Details  string
	Category sql.NullString
	Status   sql.NullString
}

func getProjectData(projectID, tenantID string) (*ProjectData, error) {
	var project ProjectData
	err := tenantManagement.DB.QueryRow(`
		SELECT title, status, category, description, complexity
		FROM st_schema.projects
		WHERE id = $1 AND tenant_id = $2
	`, projectID, tenantID).Scan(
		&project.Title,
		&project.Status,
		&project.Category,
		&project.Description,
		&project.Complexity,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching project: %w", err)
	}

	rows, err := tenantManagement.DB.Query(`
		SELECT title, details, category, status
		FROM st_schema.project_requirements
		WHERE project_id = $1 AND tenant_id = $2
	`, projectID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error fetching requirements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r RequirementData
		err := rows.Scan(&r.Title, &r.Details, &r.Category, &r.Status)
		if err != nil {
			return nil, fmt.Errorf("error scanning requirement: %w", err)
		}
		project.Requirements = append(project.Requirements, r)
	}

	return &project, nil
}

func getTemplateData(templateID, tenantID string) (*ProjectData, error) {
	var data ProjectData
	err := tenantManagement.DB.QueryRow(`
		SELECT title, 'Template' AS status, category, description, complexity
		FROM st_schema.project_templates
		WHERE id = $1 AND tenant_id = $2
	`, templateID, tenantID).Scan(
		&data.Title,
		&data.Status,
		&data.Category,
		&data.Description,
		&data.Complexity,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching template: %w", err)
	}

	data.Requirements = []RequirementData{} // no requirements for templates
	return &data, nil
}

func getCommunityTemplateData(templateID, tenantID string) (*ProjectData, error) {
	var data ProjectData
	err := tenantManagement.DB.QueryRow(`
		SELECT title, 'Community Template' AS status, category, description, complexity
		FROM st_schema.cm_project_templates
		WHERE id = $1 AND tenant_id = $2
	`, templateID, tenantID).Scan(
		&data.Title,
		&data.Status,
		&data.Category,
		&data.Description,
		&data.Complexity,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching community template: %w", err)
	}

	data.Requirements = []RequirementData{} // no requirements for community templates
	return &data, nil
}

func formatProjectInfo(data *ProjectData) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project: %s\n", data.Title))
	sb.WriteString(fmt.Sprintf("Status: %s\n", data.Status))
	sb.WriteString(fmt.Sprintf("Category: %s\n", data.Category))

	if data.Description.Valid && strings.TrimSpace(data.Description.String) != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", data.Description.String))
	}
	if data.Complexity.Valid && strings.TrimSpace(data.Complexity.String) != "" {
		sb.WriteString(fmt.Sprintf("Complexity: %s\n", data.Complexity.String))
	}

	if len(data.Requirements) > 0 {
		sb.WriteString("\nRequirements:\n")
		for _, r := range data.Requirements {
			status := "-"
			if r.Status.Valid {
				status = r.Status.String
			}
			category := "-"
			if r.Category.Valid {
				category = r.Category.String
			}

			sb.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", status, r.Title, category))
			if strings.TrimSpace(r.Details) != "" {
				sb.WriteString(fmt.Sprintf("%s\n", r.Details))
			}
		}
	}

	return sb.String()
}

func GetProjectInfo(chatCtx *models.ChatContext) (string, error) {
	var projectData *ProjectData
	var err error

	if chatCtx.ResourceGroupID == nil {
		return "", fmt.Errorf("project/template does not exists or it's ID is missing")
	}

	if chatCtx.ResourceGroupType == models.ResourceGroupProject {
		projectData, err = getProjectData(*chatCtx.ResourceGroupID, chatCtx.TenantID)
	} else if chatCtx.ResourceGroupType == models.ResourceGroupTemplate {
		projectData, err = getTemplateData(*chatCtx.ResourceGroupID, chatCtx.TenantID)
	} else if chatCtx.ResourceGroupType == models.ResourceGroupCommunity {
		projectData, err = getCommunityTemplateData(*chatCtx.ResourceGroupID, chatCtx.TenantID)
	} else {
		return "", fmt.Errorf("invalid resource type: %s", chatCtx.ResourceGroupType)
	}

	if err != nil {
		return "", err
	}

	return formatProjectInfo(projectData), nil
}

func GetProjectInfoHandler(c *gin.Context) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	chatCtx := models.ChatContext{
		UserID:   userID,
		TenantID: tenantID,
		ResourceIdentifier: models.ResourceIdentifier{
			ResourceGroupType: models.ResourceGroupProject,
			ResourceGroupID:   &projectID,
		},
	}

	projectInfo, err := GetProjectInfo(&chatCtx)
	if err != nil {
		log.Printf("Failed to retrieve project info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": models.InternalServerError})
		return
	}

	c.String(http.StatusOK, projectInfo)
}
