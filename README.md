# GoAPI

TODO: Update the readme.

This web api will be leveraged to service the `Solution Pilot` software. Backend is protected by Auth0.

## Directory Structure

## Required Environment Variables

### Database Vars

- DB_HOST
- DB_PORT
- DB_USER
- DB_PASS
- DB_NAME

### Auth0 Vars

- AUTH0_DOMAIN
- AUTH0_AUDIENCE

---

## Test app locally

Start service  
`make start`

Generate Access Token  
`make token`

Sonar Scan  
`make scan`

### Service Tests

Run test from test directory  
`make test`

If you want to make an update to code you will have to stop the service. If you want to restart the service to test your changes you will first have to kill the process that occupies port 8080.  
Find what is the process ID:

```sh
lsof -nP -iTCP:8080 -sTCP:LISTEN | awk 'NR>1 {print $2}' | xargs kill -9 | go run main.go
```

Kill the process:

```sh
kill -9
```

And again start the service:

```sh
go run main.go
```

---

## Go Tips

To install packages run:  
`go get -u packageName`

To update packages:  
`go get -u all`

To clean up packages:  
`go mod tidy`

Run tests with summary:  
`gotestsum --junitfile testresults.xml ./handlers/projects`

---

## Send request with token

```bash
curl --request POST \
  --url 'https://YOUR_DOMAIN/oauth/token' \
  --header 'content-type: application/json' \
  --data '{"grant_type":"password","username":"YOUR_USERNAME","password":"YOUR_PASSWORD","audience":"YOUR_API_IDENTIFIER","scope":"YOUR_SCOPES","client_id":"YOUR_CLIENT_ID","client_secret":"YOUR_CLIENT_SECRET"}'
```

---

## Azure & Docker

```sh
az login
az acr login --name spsaas001prodeusacr
```

Set your release tag:

```sh
RELEASE=0911
```

### Build Docker Image

```sh
docker build --platform linux/amd64 -t goapi .
docker tag goapi spsaas001prodeusacr.azurecr.io/goapi:$RELEASE
docker push spsaas001prodeusacr.azurecr.io/goapi:$RELEASE
```

Update Azure Container App:

```sh
az containerapp update \
  --name dev-goapi \
  --resource-group sp-saas-inf-001-dev-eastus-rg-000-app \
  --image spsaas001prodeusacr.azurecr.io/goapi:$RELEASE
```

---

## Logs

```kusto
ContainerAppConsoleLogs_CL 
| where ContainerAppName_s == 'dev-goapi'
| where ContainerGroupName_s == 'dev-goapi--w3chkc8-75b5c7796f-mwjjl'
| project Log_s
```

---

# API Endpoints

---

## Projects

Project management endpoints for creating, retrieving, updating, and deleting projects, as well as working with project templates and entities.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/project                         | Create a new project and default document for the user. |
| GET    | /api/project                         | Retrieve details of a specific project for the user. |
| GET    | /api/projects                        | Retrieve a list of all projects for the user. |
| PUT    | /api/project                         | Update an existing project. |
| DELETE | /api/project                         | Delete a project and its associated documents. |
| POST   | /api/projectFromTemplate             | Create a new project from a private template. |
| POST   | /api/pub/projectFromPubTemplate      | Create a new project from a public template. |
| GET    | /api/projectEntities                 | List all entities (documents, diagrams, decisions, etc.) for a given project. |

---

## Documents

Endpoints for managing documents within projects, including CRUD operations and template-based creation.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/document                        | Create a new document for a project. |
| GET    | /api/document                        | Retrieve a specific document for a project. |
| GET    | /api/documents                       | Retrieve all documents for a project. |
| PUT    | /api/document                        | Update a document for a project. |
| DELETE | /api/document                        | Delete a document from a project. |
| POST   | /api/documentTemplate                | Create a new document from a private template. |
| POST   | /api/pdt/docPubTemplate              | Create a new document from a public template. |
| POST   | /api/documentFromTemplate            | Create a new document using a private document template. |
| POST   | /api/pub/documentFromPubTemplate     | Create a new document using a public document template. |

---

## Conversations

Endpoints for managing conversations (e.g., chat sessions) within projects.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/conversation                    | Create a new conversation for a project. |
| GET    | /api/conversation                    | Retrieve a specific conversation for a project. |
| GET    | /api/conversations                   | Retrieve all conversations for a project. |
| PUT    | /api/conversation                    | Update a conversation for a project. |
| DELETE | /api/conversation                    | Delete a conversation from a project. |

---

## Diagrams

Endpoints for managing diagrams associated with projects.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/diagram                         | Create a new diagram for a project. |
| GET    | /api/diagram                         | Retrieve a specific diagram for a project. |
| GET    | /api/diagrams                        | Retrieve all diagrams for a project. |
| PUT    | /api/diagram                         | Update a diagram for a project. |
| DELETE | /api/diagram                         | Delete a diagram from a project. |

---

## Decision Support (T-Bar, PNC, SWOT, Matrix)

Endpoints for decision analysis tools: T-Bar (T-Chart), Pros & Cons, SWOT, and Decision Matrix.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| GET    | /api/tbars                           | List all T-Bar analyses for a project. |
| POST   | /api/tbar                            | Create a new T-Bar analysis. |
| GET    | /api/tbar                            | Retrieve a specific T-Bar analysis. |
| PUT    | /api/tbar                            | Update a T-Bar analysis. |
| DELETE | /api/tbar                            | Delete a T-Bar analysis. |
| POST   | /api/tbar/argument                   | Add an argument to a T-Bar option. |
| GET    | /api/tbar/arguments                  | List arguments for a T-Bar option. |
| PUT    | /api/tbar/argument                   | Update a T-Bar argument. |
| DELETE | /api/tbar/argument                   | Delete a T-Bar argument. |
| POST   | /api/pnc                             | Create a new Pros & Cons analysis. |
| PUT    | /api/pnc                             | Update a Pros & Cons analysis. |
| GET    | /api/pnc                             | Retrieve a specific Pros & Cons analysis. |
| GET    | /api/pncs                            | List all Pros & Cons analyses for a project. |
| DELETE | /api/pnc                             | Delete a Pros & Cons analysis. |
| POST   | /api/pncArgument                     | Add an argument to a Pros & Cons analysis. |
| GET    | /api/pncArguments                    | List arguments for a Pros & Cons analysis. |
| PUT    | /api/pncArgument                     | Update a Pros & Cons argument. |
| DELETE | /api/pncArgument                     | Delete a Pros & Cons argument. |
| POST   | /api/swot                            | Create a new SWOT analysis. |
| GET    | /api/swot                            | Retrieve a specific SWOT analysis. |
| GET    | /api/swots                           | List all SWOT analyses for a project. |
| PUT    | /api/swot                            | Update a SWOT analysis. |
| DELETE | /api/swot                            | Delete a SWOT analysis. |
| POST   | /api/swotArgument                    | Add an argument to a SWOT analysis. |
| GET    | /api/swotArguments                   | List arguments for a SWOT analysis. |
| PUT    | /api/swotArgument                    | Update a SWOT argument. |
| DELETE | /api/swotArgument                    | Delete a SWOT argument. |
| POST   | /api/matrix                          | Create a new decision matrix. |
| GET    | /api/matrix                          | Retrieve a specific decision matrix. |
| GET    | /api/matrixs                         | List all decision matrices for a project. |
| PUT    | /api/matrix                          | Update a decision matrix. |
| DELETE | /api/matrix                          | Delete a decision matrix. |
| POST   | /api/matrixCriteria                  | Add criteria to a decision matrix. |
| PUT    | /api/matrixCriteria                  | Update matrix criteria. |
| DELETE | /api/matrixCriteria                  | Delete matrix criteria. |
| GET    | /api/matrixCriterias                 | List all matrix criteria. |
| GET    | /api/matrixCriteria                  | Retrieve a specific matrix criteria. |
| POST   | /api/matrixConcept                   | Add a concept to a decision matrix. |
| GET    | /api/matrixConcept                   | Retrieve a specific matrix concept. |
| GET    | /api/matrixConcepts                  | List all matrix concepts. |
| PUT    | /api/matrixConcept                   | Update a matrix concept. |
| DELETE | /api/matrixConcept                   | Delete a matrix concept. |
| PUT    | /api/matrixUserRating                | Update a user's rating for a matrix. |

---

## AI Assistants & Templates

Endpoints for managing AI assistant templates, publishing, cloning, and accessing Solution Pilot and community templates.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/tenantAiTemplate                | Create a new tenant AI template. |
| GET    | /api/tenantAiTemplate                | Retrieve a specific tenant AI template. |
| GET    | /api/tenantAiTemplates               | List all tenant AI templates. |
| PUT    | /api/tenantAiTemplate                | Update a tenant AI template. |
| DELETE | /api/tenantAiTemplate                | Delete a tenant AI template. |
| PUT    | /api/tenantAiTemplate/publish        | Publish a tenant AI template to the community. |
| PUT    | /api/tenantAiTemplate/unpublish      | Unpublish a tenant AI template from the community. |
| POST   | /api/tenantAiTemplate/clone          | Clone a public AI template into the tenant's repository. |
| GET    | /api/spAiTemplate                    | Retrieve a Solution Pilot AI template. |
| GET    | /api/spAiTemplates                   | List all Solution Pilot AI templates. |
| GET    | /api/publicTemplate                  | Retrieve a public AI template (requires JWT). |
| GET    | /api/publicTemplates                 | List all public AI templates (requires JWT). |
| GET    | /api/pspAiTemplate                   | Retrieve a Solution Pilot AI template (public, no JWT). |
| GET    | /api/pspAiTemplates                  | List all Solution Pilot AI templates (public, no JWT). |
| GET    | /api/ppublicTemplate                 | Retrieve a public user AI template (public, no JWT). |
| GET    | /api/ppublicTemplates                | List all public user AI templates (public, no JWT). |

---

## Templates (Project, Document, Diagram, Components)

Endpoints for managing project, document, and diagram templates, both internal and community, as well as document components.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/projectTemplate                 | Create a new project template. |
| GET    | /api/projectTemplate                 | Retrieve a specific project template. |
| GET    | /api/projectTemplates                | List all project templates. |
| PUT    | /api/projectTemplate                 | Update a project template. |
| DELETE | /api/projectTemplate                 | Delete a project template. |
| POST   | /api/idt/documentTemplate            | Create a new internal document template. |
| GET    | /api/idt/documentTemplate            | Retrieve a specific internal document template. |
| GET    | /api/idt/documentTemplates           | List all internal document templates. |
| PUT    | /api/idt/documentTemplate            | Update an internal document template. |
| DELETE | /api/idt/documentTemplate            | Delete an internal document template. |
| POST   | /api/publicDocumentTemplate          | Create a new community document template. |
| GET    | /api/publicDocumentTemplate          | Retrieve a specific community document template. |
| GET    | /api/publicDocumentTemplates         | List all community document templates. |
| PUT    | /api/publicDocumentTemplate          | Update a community document template. |
| DELETE | /api/publicDocumentTemplate          | Delete a community document template. |
| POST   | /api/idt/diagramTemplate             | Create a new internal diagram template. |
| GET    | /api/idt/diagramTemplate             | Retrieve a specific internal diagram template. |
| GET    | /api/idt/diagramTemplates            | List all internal diagram templates. |
| PUT    | /api/idt/diagramTemplate             | Update an internal diagram template. |
| DELETE | /api/idt/diagramTemplate             | Delete an internal diagram template. |
| POST   | /api/publicDiagramTemplate           | Create a new community diagram template. |
| GET    | /api/publicDiagramTemplate           | Retrieve a specific community diagram template. |
| GET    | /api/publicDiagramTemplates          | List all community diagram templates. |
| PUT    | /api/publicDiagramTemplate           | Update a community diagram template. |
| DELETE | /api/publicDiagramTemplate           | Delete a community diagram template. |
| GET    | /api/dcm/component                   | Retrieve a document component. |
| GET    | /api/dcm/components                  | List all document components. |
| GET    | /api/dcm/favoriteComponents          | List favorite document components. |
| POST   | /api/dcm/pinComponent                | Pin a document component. |
| POST   | /api/dcm/unpinComponent              | Unpin a document component. |
| POST   | /api/publishProjectTemplate          | Publish a project template to the community. |
| POST   | /api/unpublishProjectTemplate        | Unpublish a project template from the community. |

---

## Public Template Routers

Endpoints for accessing public project, document, and diagram templates, including website-specific APIs.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| GET    | /api/publicProjectTemplate           | Retrieve a public project template. |
| GET    | /api/publicProjectTemplates          | List all public project templates. |
| GET    | /api/publicProjectDocumentTemplate   | Retrieve a public project document template. |
| GET    | /api/publicProjectDiagramTemplate    | Retrieve a public project diagram template. |
| PUT    | /api/publicProjectTemplate           | Update a public project template. |
| GET    | /api/pub/publicProjectTemplate       | Retrieve a public project template (website, no JWT). |
| GET    | /api/pub/publicProjectTemplates      | List all public project templates (website, no JWT). |
| GET    | /api/pub/publicProjectTemplatesPag   | Paginated list of public project templates (website, no JWT). |
| GET    | /api/pub/publicProjectDocumentTemplate | Retrieve a public project document template (website, no JWT). |
| GET    | /api/pub/publicProjectDiagramTemplate | Retrieve a public project diagram template (website, no JWT). |
| POST   | /api/clonePublicProjectTemplate      | Clone a public project template into the tenant's repository. |

---

## Comments

Endpoints for managing public comments on templates or projects.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/publicComment                   | Post a new public comment. |
| GET    | /api/publicComments                  | List all public comments. |
| PUT    | /api/publicComment                   | Update a public comment. |
| DELETE | /api/publicComment                   | Delete a public comment. |

---

## Cloud Resource Management

Endpoints for managing cloud resources and credentials (e.g., Azure).

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/cloud/resources                 | Get Azure resources by tag. |
| GET    | /api/cloud/credentials               | Retrieve tenant cloud credentials. |
| POST   | /api/cloud/credentials               | Create a new tenant cloud credential. |
| DELETE | /api/cloud/credentials               | Delete a tenant cloud credential. |

---

## Tiptap Collaboration

Endpoints for collaborative editing and real-time features.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| GET    | /api/tiptap/collab                   | Real-time collaboration endpoint for Tiptap editor. |

---

## Tenant & User Management

Endpoints for managing users, tenants, members, and invitations.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| GET    | /api/user                            | Retrieve the current user's profile. |
| PATCH  | /api/user                            | Update the current user's profile. |
| POST   | /api/m/NewAzOiAccount                | Create a new tenant (admin only). |
| POST   | /api/paymentSession                  | Create a Stripe payment session (admin only). |
| GET    | /api/tenant                          | Retrieve tenant details and members. |
| PUT    | /api/tenant                          | Update tenant details (admin only). |
| PUT    | /api/tenant/members/:member_id       | Update a tenant member (admin only). |
| DELETE | /api/tenant/members/:member_id       | Remove a tenant member (admin only). |
| GET    | /api/tenant/invitations              | List all tenant invitations (admin only). |
| POST   | /api/tenant/invitations              | Invite users to the tenant (admin only). |
| DELETE | /api/tenant/invitations/:invitation_id | Delete a tenant invitation (admin only). |

---

## AI Prompt & Text Utilities

Endpoints for AI-powered text manipulation and OpenAI prompt streaming.

| Method | Endpoint | Description |
|--------|---------------------------------------|----------------------------------------------------------------------------------------------|
| POST   | /api/newPrompt                       | Create a new OpenAI prompt (Solution Pilot agent or custom). |
| GET    | /api/newChatHistory                  | Retrieve chat history for a prompt. |
| GET    | /api/newAiStream                     | Stream OpenAI responses in real time. |
| POST   | /api/text/simplify                   | Simplify text using AI. |
| POST   | /api/text/fix-spelling-and-grammar   | Fix spelling and grammar in text using AI. |
| POST   | /api/text/shorten                    | Shorten text using AI. |
| POST   | /api/text/extend                     | Extend text using AI. |
| POST   | /api/text/adjust-tone                | Adjust the tone of text using AI. |
| POST   | /api/text/tldr                       | Summarize text using AI. |
| POST   | /api/text/prompt                     | Generate text from a prompt using AI. |
| POST   | /api/text/autocomplete               | Autocomplete text using AI. |

---