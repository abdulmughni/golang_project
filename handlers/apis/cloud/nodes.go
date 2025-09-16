package cloud

type AzureNodeItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

type AzureNodeCategory struct {
	Category     string          `json:"category"`
	CategoryIcon string          `json:"categoryIcon"`
	Items        []AzureNodeItem `json:"items"`
}

var AzureNodes = []AzureNodeCategory{
	{
		Category:     "Web",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-service-plans.svg",
		Items: []AzureNodeItem{
			{
				Name: "api-center",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/api-center.svg",
				Type: "Microsoft.ApiCenter/services",
			},
			{
				Name: "api-connections",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/api-connections.svg",
				Type: "Microsoft.Web/connections",
			},
			{
				Name: "api-management-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/api-management-services.svg",
				Type: "Microsoft.ApiManagement/service",
			},
			{
				Name: "app-service-certificates",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-service-certificates.svg",
				Type: "Microsoft.CertificateRegistration/certificateOrders",
			},
			{
				Name: "app-service-domains",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-service-domains.svg",
				Type: "Microsoft.DomainRegistration/domains",
			},
			{
				Name: "app-service-environments",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-service-environments.svg",
				Type: "Microsoft.Web/hostingEnvironments",
			},
			{
				Name: "app-service-plans",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-service-plans.svg",
				Type: "Microsoft.Web/serverFarms",
			},
			{
				Name: "app-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/app-services.svg",
				Type: "Microsoft.Web/sites",
			},
			{
				Name: "function-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/function-apps.svg",
				Type: "Microsoft.Web/sites",
			},
			{
				Name: "azure-media-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/azure-media-service.svg",
				Type: "Microsoft.Media/mediaservices",
			},
			{
				Name: "azure-spring-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/azure-spring-apps.svg",
				Type: "Microsoft.AppPlatform/Spring",
			},
			{
				Name: "cognitive-search",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/cognitive-search.svg",
				Type: "Microsoft.Search/searchServices",
			},
			{
				Name: "cognitive-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/cognitive-services.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "front-door-and-cdn-profiles",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/front-door-and-cdn-profiles.svg",
				Type: "Microsoft.Cdn/profiles",
			},
			{
				Name: "notification-hub-namespaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/notification-hub-namespaces.svg",
				Type: "Microsoft.NotificationHubs/namespaces",
			},
			{
				Name: "power-platform",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/power-platform.svg",
				Type: "Microsoft.PowerPlatform/enterprisePolicies",
			},
			{
				Name: "signalr",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/signalr.svg",
				Type: "Microsoft.SignalRService/SignalR",
			},
			{
				Name: "static-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/web/static-apps.svg",
				Type: "Microsoft.Web/staticSites",
			},
		},
	},
	{
		Category:     "Containers",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/kubernetes-services.svg",
		Items: []AzureNodeItem{
			{
				Name: "kubernetes-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/kubernetes-services.svg",
				Type: "Microsoft.ContainerService/managedClusters",
			},
			{
				Name: "container-apps-environments",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/container-apps-environments.svg",
				Type: "Microsoft.App/managedEnvironments",
			},
			{
				Name: "container-instances",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/container-instances.svg",
				Type: "Microsoft.ContainerInstance/containerGroups",
			},
			{
				Name: "container-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/worker-container-app.svg",
				Type: "Microsoft.App/containerApps",
			},
			{
				Name: "container-registries",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/container-registries.svg",
				Type: "Microsoft.ContainerRegistry/registries",
			},
			{
				Name: "azure-red-hat-openshift",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/azure-red-hat-openshift.svg",
				Type: "Microsoft.RedHatOpenShift/openShiftClusters",
			},
			{
				Name: "batch-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/batch-accounts.svg",
				Type: "Microsoft.Batch/batchAccounts",
			},
			{
				Name: "service-fabric-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/containers/service-fabric-clusters.svg",
				Type: "Microsoft.ServiceFabric/clusters",
			},
		},
	},
	{
		Category:     "Compute",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/virtual-machine.svg",
		Items: []AzureNodeItem{
			{
				Name: "virtual-machine",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/virtual-machine.svg",
				Type: "Microsoft.Compute/virtualMachines",
			},
			{
				Name: "vm-scale-sets",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/vm-scale-sets.svg",
				Type: "Microsoft.Compute/virtualMachineScaleSets",
			},
			{
				Name: "automanaged-vm",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/automanaged-vm.svg",
				Type: "Microsoft.Compute/virtualMachines",
			},
			{
				Name: "availability-sets",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/availability-sets.svg",
				Type: "Microsoft.Compute/availabilitySets",
			},
			{
				Name: "azure-compute-galleries",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/azure-compute-galleries.svg",
				Type: "Microsoft.Compute/galleries",
			},
			{
				Name: "batch-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/batch-accounts.svg",
				Type: "Microsoft.Batch/batchAccounts",
			},
			{
				Name: "disk-encryption-sets",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/disk-encryption-sets.svg",
				Type: "Microsoft.Compute/diskEncryptionSets",
			},
			{
				Name: "disks-snapshots",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/disks-snapshots.svg",
				Type: "Microsoft.Compute/snapshots",
			},
			{
				Name: "disks",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/disks.svg",
				Type: "Microsoft.Compute/disks",
			},
			{
				Name: "host-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/host-groups.svg",
				Type: "Microsoft.Compute/hostGroups",
			},
			{
				Name: "host-pools",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/host-pools.svg",
				Type: "Microsoft.DesktopVirtualization/hostPools",
			},
			{
				Name: "hosts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/hosts.svg",
				Type: "Microsoft.Compute/hosts",
			},
			{
				Name: "image-definitions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/image-definitions.svg",
				Type: "Microsoft.Compute/images",
			},
			{
				Name: "image-templates",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/image-templates.svg",
				Type: "Microsoft.VirtualMachineImages/imageTemplates",
			},
			{
				Name: "image-versions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/image-versions.svg",
				Type: "Microsoft.Compute/images",
			},
			{
				Name: "images",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/images.svg",
				Type: "Microsoft.Compute/images",
			},
			{
				Name: "maintenance-configuration",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/maintenance-configuration.svg",
				Type: "Microsoft.Maintenance/configurations",
			},
			{
				Name: "managed-service-fabric",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/managed-service-fabric.svg",
				Type: "Microsoft.ServiceFabric/managedClusters",
			},
			{
				Name: "mesh-applications",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/mesh-applications.svg",
				Type: "Microsoft.ServiceFabricMesh/applications",
			},
			{
				Name: "metrics-advisor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/metrics-advisor.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "restore-points-collections",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/restore-points-collections.svg",
				Type: "Microsoft.Compute/restorePointCollections",
			},
			{
				Name: "restore-points",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/restore-points.svg",
				Type: "Microsoft.Compute/restorePoints",
			},
			{
				Name: "service-fabric-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/service-fabric-clusters.svg",
				Type: "Microsoft.ServiceFabric/clusters",
			},
			{
				Name: "shared-image-galleries",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/shared-image-galleries.svg",
				Type: "Microsoft.Compute/galleries",
			},
			{
				Name: "workspaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/compute/workspaces.svg",
				Type: "Microsoft.DesktopVirtualization/workspaces",
			},
		},
	},
	{
		Category:     "Networking",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/virtual-networks.svg",
		Items: []AzureNodeItem{
			{
				Name: "virtual-networks",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/virtual-networks.svg",
				Type: "Microsoft.Network/virtualNetworks",
			},
			{
				Name: "subnet",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/subnet.svg",
				Type: "Microsoft.Network/virtualNetworks/subnets",
			},
			{
				Name: "firewalls",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/firewalls.svg",
				Type: "Microsoft.Network/azureFirewalls",
			},
			{
				Name: "azure-firewall-manager",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/azure-firewall-manager.svg",
				Type: "Microsoft.Network/firewallPolicies",
			},
			{
				Name: "azure-firewall-policy",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/azure-firewall-policy.svg",
				Type: "Microsoft.Network/firewallPolicies",
			},
			{
				Name: "front-door-and-cdn-profiles",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/front-door-and-cdn-profiles.svg",
				Type: "Microsoft.Cdn/profiles",
			},
			{
				Name: "cdn-profiles",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/cdn-profiles.svg",
				Type: "Microsoft.Cdn/profiles",
			},
			{
				Name: "application-gateways",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/application-gateways.svg",
				Type: "Microsoft.Network/applicationGateways",
			},
			{
				Name: "application-gateway-containers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/application-gateway-containers.svg",
				Type: "Microsoft.Network/applicationGateways",
			},
			{
				Name: "load-balancers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/load-balancers.svg",
				Type: "Microsoft.Network/loadBalancers",
			},
			{
				Name: "load-balancer-hub",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/load-balancer-hub.svg",
				Type: "Microsoft.Network/loadBalancers",
			},
			{
				Name: "dns-zones",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/dns-zones.svg",
				Type: "Microsoft.Network/privateDnsZones",
			},
			{
				Name: "dns-multistack",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/dns-multistack.svg",
				Type: "Microsoft.Network/dnsZones",
			},
			{
				Name: "dns-security-policy",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/dns-security-policy.svg",
				Type: "Microsoft.Network/dnsZones",
			},
			{
				Name: "traffic-manager-profiles",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/traffic-manager-profiles.svg",
				Type: "Microsoft.Network/trafficManagerProfiles",
			},
			{
				Name: "virtual-network-gateways",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/virtual-network-gateways.svg",
				Type: "Microsoft.Network/virtualNetworkGateways",
			},
			{
				Name: "virtual-router",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/virtual-router.svg",
				Type: "Microsoft.Network/virtualRouters",
			},
			{
				Name: "virtual-wan-hub",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/virtual-wan-hub.svg",
				Type: "Microsoft.Network/virtualWans",
			},
			{
				Name: "virtual-wans",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/virtual-wans.svg",
				Type: "Microsoft.Network/virtualWans",
			},
			{
				Name: "expressroute-circuits",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/expressroute-circuits.svg",
				Type: "Microsoft.Network/expressRouteCircuits",
			},
			{
				Name: "local-network-gateways",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/local-network-gateways.svg",
				Type: "Microsoft.Network/localNetworkGateways",
			},
			{
				Name: "bastions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/bastions.svg",
				Type: "Microsoft.Network/bastionHosts",
			},
			{
				Name: "connected-cache",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/connected-cache.svg",
				Type: "Microsoft.Network/connectedCaches",
			},
			{
				Name: "connections",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/connections.svg",
				Type: "Microsoft.Network/connections",
			},
			{
				Name: "ddos-protection-plans",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/ddos-protection-plans.svg",
				Type: "Microsoft.Network/ddosProtectionPlans",
			},
			{
				Name: "ip-address-manager",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/ip-address-manager.svg",
				Type: "Microsoft.Network/publicIPAddresses",
			},
			{
				Name: "ip-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/ip-groups.svg",
				Type: "Microsoft.Network/ipGroups",
			},
			{
				Name: "azure-communications-gateway",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/azure-communications-gateway.svg",
				Type: "Microsoft.Network/azureCommunicationGateways",
			},
			{
				Name: "nat",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/nat.svg",
				Type: "Microsoft.Network/natGateways",
			},
			{
				Name: "network-interfaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/network-interfaces.svg",
				Type: "Microsoft.Network/networkInterfaces",
			},
			{
				Name: "network-security-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/network-security-groups.svg",
				Type: "Microsoft.Network/networkSecurityGroups",
			},
			{
				Name: "network-watcher",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/network-watcher.svg",
				Type: "Microsoft.Network/networkWatchers",
			},
			{
				Name: "on-premises-data-gateways",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/on-premises-data-gateways.svg",
				Type: "Microsoft.Network/onPremisesDataGateways",
			},
			{
				Name: "private-link-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/private-link-service.svg",
				Type: "Microsoft.Network/privateLinkServices",
			},
			{
				Name: "private-link-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/private-link-services.svg",
				Type: "Microsoft.Network/privateLinkServices",
			},
			{
				Name: "private-link",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/private-link.svg",
				Type: "Microsoft.Network/privateLinkServices",
			},
			{
				Name: "proximity-placement-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/proximity-placement-groups.svg",
				Type: "Microsoft.Compute/proximityPlacementGroups",
			},
			{
				Name: "private-endpoints",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/private-endpoints.svg",
				Type: "Microsoft.Network/privateEndpoints",
			},
			{
				Name: "public-ip-addresses",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/public-ip-addresses.svg",
				Type: "Microsoft.Network/publicIPAddresses",
			},
			{
				Name: "public-ip-prefixes",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/public-ip-prefixes.svg",
				Type: "Microsoft.Network/publicIPPrefixes",
			},
			{
				Name: "resource-management-private-link",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/resource-management-private-link.svg",
				Type: "Microsoft.Network/privateLinkServices",
			},
			{
				Name: "route-filters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/route-filters.svg",
				Type: "Microsoft.Network/routeFilters",
			},
			{
				Name: "route-tables",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/route-tables.svg",
				Type: "Microsoft.Network/routeTables",
			},
			{
				Name: "service-endpoint-policies",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/service-endpoint-policies.svg",
				Type: "Microsoft.Network/serviceEndpointPolicies",
			},
			{
				Name: "spot-vm",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/spot-vm.svg",
				Type: "Microsoft.Compute/virtualMachines",
			},
			{
				Name: "spot-vmss",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/spot-vmss.svg",
				Type: "Microsoft.Compute/virtualMachineScaleSets",
			},
			{
				Name: "web-application-firewall-policies(waf)",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/web-application-firewall-policies(waf).svg",
				Type: "Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies",
			},
			{
				Name: "atm-multistack",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/networking/atm-multistack.svg",
				Type: "Microsoft.Network/virtualNetworks",
			},
		},
	},
	{
		Category:     "Databases",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-server.svg",
		Items: []AzureNodeItem{
			{
				Name: "azure-cosmos-db",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-cosmos-db.svg",
				Type: "Microsoft.DocumentDB/databaseAccounts",
			},
			{
				Name: "azure-data-explorer-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-data-explorer-clusters.svg",
				Type: "Microsoft.Kusto/clusters",
			},
			{
				Name: "azure-database-mariadb-server",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-database-mariadb-server.svg",
				Type: "Microsoft.DBforMariaDB/servers",
			},
			{
				Name: "azure-database-migration-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-database-migration-services.svg",
				Type: "Microsoft.DataMigration/services",
			},
			{
				Name: "azure-database-mysql-server",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-database-mysql-server.svg",
				Type: "Microsoft.DBforMySQL/servers",
			},
			{
				Name: "azure-database-postgresql-server-group",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-database-postgresql-server-group.svg",
				Type: "Microsoft.DBforPostgreSQL/serverGroups",
			},
			{
				Name: "azure-database-postgresql-server",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-database-postgresql-server.svg",
				Type: "Microsoft.DBforPostgreSQL/flexibleServers",
			},
			{
				Name: "azure-purview-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-purview-accounts.svg",
				Type: "Microsoft.Purview/accounts",
			},
			{
				Name: "azure-sql-edge",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-sql-edge.svg",
				Type: "Microsoft.Sql/servers",
			},
			{
				Name: "azure-sql-server-stretch-databases",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-sql-server-stretch-databases.svg",
				Type: "Microsoft.Sql/servers/databases",
			},
			{
				Name: "azure-sql-vm",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-sql-vm.svg",
				Type: "Microsoft.SqlVirtualMachine/sqlVirtualMachines",
			},
			{
				Name: "azure-sql",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-sql.svg",
				Type: "Microsoft.Sql/servers",
			},
			{
				Name: "azure-synapse-analytics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/azure-synapse-analytics.svg",
				Type: "Microsoft.Synapse/workspaces",
			},
			{
				Name: "cache-redis",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/cache-redis.svg",
				Type: "Microsoft.Cache/Redis",
			},
			{
				Name: "data-factories",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/data-factories.svg",
				Type: "Microsoft.DataFactory/factories",
			},
			{
				Name: "elastic-job-agents",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/elastic-job-agents.svg",
				Type: "Microsoft.Sql/servers/elasticJobAgents",
			},
			{
				Name: "instance-pools",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/instance-pools.svg",
				Type: "Microsoft.Sql/instancePools",
			},
			{
				Name: "managed-database",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/managed-database.svg",
				Type: "Microsoft.Sql/managedInstances/databases",
			},
			{
				Name: "oracle-database",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/oracle-database.svg",
				Type: "Microsoft.Oracle/servers",
			},
			{
				Name: "sql-data-warehouses",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-data-warehouses.svg",
				Type: "Microsoft.Sql/servers/databases",
			},
			{
				Name: "sql-database",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-database.svg",
				Type: "Microsoft.Sql/servers/databases",
			},
			{
				Name: "sql-elastic-pools",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-elastic-pools.svg",
				Type: "Microsoft.Sql/servers/elasticPools",
			},
			{
				Name: "sql-managed-instance",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-managed-instance.svg",
				Type: "Microsoft.Sql/managedInstances",
			},
			{
				Name: "sql-server-registries",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-server-registries.svg",
				Type: "Microsoft.Sql/servers",
			},
			{
				Name: "sql-server",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/sql-server.svg",
				Type: "Microsoft.Sql/servers",
			},
			{
				Name: "ssis-lift-and-shift-ir",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/ssis-lift-and-shift-ir.svg",
				Type: "Microsoft.DataFactory/factories/integrationRuntimes",
			},
			{
				Name: "virtual-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/databases/virtual-clusters.svg",
				Type: "Microsoft.Sql/virtualClusters",
			},
		},
	},
	{
		Category:     "Storage",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/storage-accounts.svg",
		Items: []AzureNodeItem{
			{
				Name: "azure-databox-gateway",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/azure-databox-gateway.svg",
				Type: "Microsoft.DataBox/gateways",
			},
			{
				Name: "azure-fileshares",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/azure-fileshares.svg",
				Type: "Microsoft.Storage/storageAccounts/fileServices",
			},
			{
				Name: "azure-hcp-cache",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/azure-hcp-cache.svg",
				Type: "Microsoft.HCP/cache",
			},
			{
				Name: "azure-netapp-files",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/azure-netapp-files.svg",
				Type: "Microsoft.NetApp/netAppAccounts",
			},
			{
				Name: "azure-stack-edge",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/azure-stack-edge.svg",
				Type: "Microsoft.DataBoxEdge/dataBoxEdgeDevices",
			},
			{
				Name: "data-box",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/data-box.svg",
				Type: "Microsoft.DataBox/dataBoxes",
			},
			{
				Name: "data-lake-storage-gen1",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/data-lake-storage-gen1.svg",
				Type: "Microsoft.DataLakeStore/accounts",
			},
			{
				Name: "data-share-invitations",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/data-share-invitations.svg",
				Type: "Microsoft.DataShare/invitations",
			},
			{
				Name: "data-shares",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/data-shares.svg",
				Type: "Microsoft.DataShare/accounts",
			},
			{
				Name: "import-export-jobs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/import-export-jobs.svg",
				Type: "Microsoft.ImportExport/jobs",
			},
			{
				Name: "recovery-services-vaults",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/recovery-services-vaults.svg",
				Type: "Microsoft.RecoveryServices/vaults",
			},
			{
				Name: "storage-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/storage-accounts.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "storage-explorer",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/storage-explorer.svg",
				Type: "Microsoft.StorageExplorer/explorers",
			},
			{
				Name: "storage-sync-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/storage-sync-services.svg",
				Type: "Microsoft.StorageSync/storageSyncServices",
			},
			{
				Name: "storsimple-data-managers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/storsimple-data-managers.svg",
				Type: "Microsoft.StorSimple/managers",
			},
			{
				Name: "storsimple-device-managers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/storage/storsimple-device-managers.svg",
				Type: "Microsoft.StorSimple/managers",
			},
		},
	},
	{
		Category:     "Identity",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-domain-services.svg",
		Items: []AzureNodeItem{
			{
				Name: "active-directory-connect-health",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/active-directory-connect-health.svg",
				Type: "Microsoft.AAD/connectHealth",
			},
			{
				Name: "administrative-units",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/administrative-units.svg",
				Type: "Microsoft.AAD/administrativeUnits",
			},
			{
				Name: "api-proxy",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/api-proxy.svg",
				Type: "Microsoft.AAD/apiProxies",
			},
			{
				Name: "app-registrations",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/app-registrations.svg",
				Type: "Microsoft.AAD/appRegistrations",
			},
			{
				Name: "azure-ad-b2c",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/azure-ad-b2c.svg",
				Type: "Microsoft.AAD/b2cDirectories",
			},
			{
				Name: "azure-information-protection",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/azure-information-protection.svg",
				Type: "Microsoft.InformationProtection/policies",
			},
			{
				Name: "enterprise-applications",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/enterprise-applications.svg",
				Type: "Microsoft.AAD/enterpriseApplications",
			},
			{
				Name: "entra-connect",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-connect.svg",
				Type: "Microsoft.Entra/connectors",
			},
			{
				Name: "entra-domain-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-domain-services.svg",
				Type: "Microsoft.Entra/domainServices",
			},
			{
				Name: "entra-global-secure-access",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-global-secure-access.svg",
				Type: "Microsoft.Entra/globalSecureAccess",
			},
			{
				Name: "entra-id-protection",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-id-protection.svg",
				Type: "Microsoft.Entra/idProtection",
			},
			{
				Name: "entra-identity-custom-roles",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-identity-custom-roles.svg",
				Type: "Microsoft.Entra/customRoles",
			},
			{
				Name: "entra-identity-licenses",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-identity-licenses.svg",
				Type: "Microsoft.Entra/identityLicenses",
			},
			{
				Name: "entra-identity-roles-and-administrators",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-identity-roles-and-administrators.svg",
				Type: "Microsoft.Entra/rolesAndAdministrators",
			},
			{
				Name: "entra-internet-access",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-internet-access.svg",
				Type: "Microsoft.Entra/internetAccess",
			},
			// {
			// 	Name: "entra-managed-identities",
			// 	URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-managed-identities.svg",
			// 	Type: "Microsoft.ManagedIdentity/userAssignedIdentities",
			// },
			{
				Name: "entra-private-access",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-private-access.svg",
				Type: "Microsoft.Entra/privateAccess",
			},
			{
				Name: "entra-privleged-identity-management",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-privleged-identity-management.svg",
				Type: "Microsoft.Entra/privilegedIdentityManagement",
			},
			{
				Name: "entra-verified-id",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/entra-verified-id.svg",
				Type: "Microsoft.Entra/verifiedId",
			},
			{
				Name: "external-identities",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/external-identities.svg",
				Type: "Microsoft.AAD/externalIdentities",
			},
			{
				Name: "groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/groups.svg",
				Type: "Microsoft.AAD/groups",
			},
			{
				Name: "identity-governance",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/identity-governance.svg",
				Type: "Microsoft.AAD/identityGovernance",
			},
			{
				Name: "managed-identities",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/managed-identities.svg",
				Type: "Microsoft.ManagedIdentity/userAssignedIdentities",
			},
			{
				Name: "multi-factor-authentication",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/multi-factor-authentication.svg",
				Type: "Microsoft.AAD/multiFactorAuth",
			},
			{
				Name: "security",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/security.svg",
				Type: "Microsoft.AAD/security",
			},
			{
				Name: "tenant-properties",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/tenant-properties.svg",
				Type: "Microsoft.AAD/tenantProperties",
			},
			{
				Name: "user-settings",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/user-settings.svg",
				Type: "Microsoft.AAD/userSettings",
			},
			{
				Name: "users",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/users.svg",
				Type: "Microsoft.AAD/users",
			},
			{
				Name: "verifiable-credentials",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/verifiable-credentials.svg",
				Type: "Microsoft.VerifiableCredentials/credentials",
			},
			{
				Name: "verification-as-a-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/identity/verification-as-a-service.svg",
				Type: "Microsoft.Verification/verificationServices",
			},
		},
	},
	{
		Category:     "AI Services",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/ai-studio.svg",
		Items: []AzureNodeItem{
			{
				Name: "ai-studio",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/ai-studio.svg",
				Type: "Microsoft.MachineLearningServices/workspaces",
			},
			{
				Name: "azure-openai",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/azure-openai.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "batch-ai",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/batch-ai.svg",
				Type: "Microsoft.BatchAI/workspaces",
			},
			{
				Name: "bonsai",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/bonsai.svg",
				Type: "Microsoft.Bonsai/accounts",
			},
			{
				Name: "bot-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/bot-services.svg",
				Type: "Microsoft.BotService/botServices",
			},
			{
				Name: "cognitive-search",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/cognitive-search.svg",
				Type: "Microsoft.Search/searchServices",
			},
			{
				Name: "cognitive-services-decisions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/cognitive-services-decisions.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "cognitive-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/cognitive-services.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "computer-vision",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/computer-vision.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "content-moderators",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/content-moderators.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "content-safety",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/content-safety.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "custom-vision",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/custom-vision.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "face-apis",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/face-apis.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "form-recognizers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/form-recognizers.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "genomics-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/genomics-accounts.svg",
				Type: "Microsoft.Genomics/accounts",
			},
			{
				Name: "genomics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/genomics.svg",
				Type: "Microsoft.Genomics/accounts",
			},
			{
				Name: "immersive-readers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/immersive-readers.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "language-understanding",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/language-understanding.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "language",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/language.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "metrics-advisor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/metrics-advisor.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "personalizers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/personalizers.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "azure-experimentation-studio",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/azure-experimentation-studio.svg",
				Type: "Microsoft.MachineLearningServices/workspaces",
			},
			{
				Name: "qna-makers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/qna-makers.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "serverless-search",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/serverless-search.svg",
				Type: "Microsoft.Search/searchServices",
			},
			{
				Name: "speech-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/speech-services.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "translator-text",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/translator-text.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "azure-applied-ai-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/azure-applied-ai-services.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
			{
				Name: "azure-object-understanding",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/ai_services/azure-object-understanding.svg",
				Type: "Microsoft.CognitiveServices/accounts",
			},
		},
	},
	{
		Category:     "Machine Learning Services",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/machine_learning_services/machine-learning.svg",
		Items: []AzureNodeItem{
			{
				Name: "machine-learning-studio-(classic)-web-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/machine_learning_services/machine-learning-studio-(classic)-web-services.svg",
				Type: "Microsoft.MachineLearning/webServices",
			},
			{
				Name: "machine-learning-studio-web-service-plans",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/machine_learning_services/machine-learning-studio-web-service-plans.svg",
				Type: "Microsoft.MachineLearning/webServicePlans",
			},
			{
				Name: "machine-learning-studio-workspaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/machine_learning_services/machine-learning-studio-workspaces.svg",
				Type: "Microsoft.MachineLearningServices/workspaces",
			},
			{
				Name: "machine-learning",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/machine_learning_services/machine-learning.svg",
				Type: "Microsoft.MachineLearningServices/workspaces",
			},
		},
	},
	{
		Category:     "Analytics",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/log-analytics-workspaces.svg",
		Items: []AzureNodeItem{
			{
				Name: "analysis-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/analysis-services.svg",
				Type: "Microsoft.AnalysisServices/servers",
			},
			{
				Name: "azure-data-explorer-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/azure-data-explorer-clusters.svg",
				Type: "Microsoft.Kusto/clusters",
			},
			{
				Name: "azure-databricks",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/azure-databricks.svg",
				Type: "Microsoft.Databricks/workspaces",
			},
			{
				Name: "azure-synapse-analytics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/azure-synapse-analytics.svg",
				Type: "Microsoft.Synapse/workspaces",
			},
			{
				Name: "azure-workbooks",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/azure-workbooks.svg",
				Type: "Microsoft.Insights/workbooks",
			},
			{
				Name: "data-factories",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/data-factories.svg",
				Type: "Microsoft.DataFactory/factories",
			},
			{
				Name: "data-lake-analytics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/data-lake-analytics.svg",
				Type: "Microsoft.DataLakeAnalytics/accounts",
			},
			{
				Name: "data-lake-store-gen1",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/data-lake-store-gen1.svg",
				Type: "Microsoft.DataLakeStore/accounts",
			},
			{
				Name: "endpoint-analytics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/endpoint-analytics.svg",
				Type: "Microsoft.AnalysisServices/servers",
			},
			{
				Name: "event-hub-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/event-hub-clusters.svg",
				Type: "Microsoft.EventHub/clusters",
			},
			{
				Name: "event-hubs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/event-hubs.svg",
				Type: "Microsoft.EventHub/namespaces",
			},
			{
				Name: "hd-insight-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/hd-insight-clusters.svg",
				Type: "Microsoft.HDInsight/clusters",
			},
			{
				Name: "log-analytics-workspaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/log-analytics-workspaces.svg",
				Type: "Microsoft.OperationalInsights/workspaces",
			},
			{
				Name: "power-bi-embedded",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/power-bi-embedded.svg",
				Type: "Microsoft.PowerBIDedicated/capacities",
			},
			{
				Name: "power-platform",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/power-platform.svg",
				Type: "Microsoft.PowerPlatform/enterprisePolicies",
			},
			{
				Name: "private-link-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/private-link-services.svg",
				Type: "Microsoft.Network/privateLinkServices",
			},
			{
				Name: "stream-analytics-jobs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/analytics/stream-analytics-jobs.svg",
				Type: "Microsoft.StreamAnalytics/streamingJobs",
			},
		},
	},
	{
		Category:     "Azure DevOps",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/azure-devops.svg",
		Items: []AzureNodeItem{
			{
				Name: "api-connections",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/api-connections.svg",
				Type: "Microsoft.Web/connections",
			},
			{
				Name: "api-management-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/api-management-services.svg",
				Type: "Microsoft.ApiManagement/service",
			},
			{
				Name: "azure-devops",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/azure-devops.svg",
				Type: "Microsoft.DevOps/organizations",
			},
			{
				Name: "change-analysis",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/change-analysis.svg",
				Type: "Microsoft.ChangeAnalysis/changeAnalysis",
			},
			{
				Name: "cloudtest",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/cloudtest.svg",
				Type: "Microsoft.CloudTest/testPlans",
			},
			{
				Name: "code-optimization",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/code-optimization.svg",
				Type: "Microsoft.CodeOptimization/optimization",
			},
			{
				Name: "devops-starter",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/devops-starter.svg",
				Type: "Microsoft.DevOps/starter",
			},
			{
				Name: "devtest-labs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/devtest-labs.svg",
				Type: "Microsoft.DevTestLab/labs",
			},
			{
				Name: "lab-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/lab-accounts.svg",
				Type: "Microsoft.LabServices/labAccounts",
			},
			{
				Name: "lab-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/lab-services.svg",
				Type: "Microsoft.LabServices/labs",
			},
			{
				Name: "load-testing",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_devops/load-testing.svg",
				Type: "Microsoft.LoadTest/loadTests",
			},
		},
	},
	{
		Category:     "General",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/backlog.svg",
		Items: []AzureNodeItem{
			{
				Name: "all-resources",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/all-resources.svg",
				Type: "Microsoft.Resources/resources",
			},
			{
				Name: "backlog",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/backlog.svg",
				Type: "Microsoft.DevOps/backlog",
			},
			{
				Name: "biz-talk",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/biz-talk.svg",
				Type: "Microsoft.BizTalk/services",
			},
			{
				Name: "blob-block",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/blob-block.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "blob-page",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/blob-page.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "branch",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/branch.svg",
				Type: "Microsoft.DevOps/branches",
			},
			{
				Name: "browser",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/browser.svg",
				Type: "Microsoft.Web/browsers",
			},
			{
				Name: "bug",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/bug.svg",
				Type: "Microsoft.DevOps/bugs",
			},
			{
				Name: "builds",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/builds.svg",
				Type: "Microsoft.DevOps/builds",
			},
			{
				Name: "cache",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cache.svg",
				Type: "Microsoft.Cache/Redis",
			},
			{
				Name: "code",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/code.svg",
				Type: "Microsoft.DevOps/code",
			},
			{
				Name: "commit",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/commit.svg",
				Type: "Microsoft.DevOps/commits",
			},
			{
				Name: "controls-horizontal",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/controls-horizontal.svg",
				Type: "Microsoft.DevOps/controls",
			},
			{
				Name: "controls",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/controls.svg",
				Type: "Microsoft.DevOps/controls",
			},
			{
				Name: "cost-alerts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cost-alerts.svg",
				Type: "Microsoft.CostManagement/alerts",
			},
			{
				Name: "cost-analysis",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cost-analysis.svg",
				Type: "Microsoft.CostManagement/analysis",
			},
			{
				Name: "cost-budgets",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cost-budgets.svg",
				Type: "Microsoft.CostManagement/budgets",
			},
			{
				Name: "cost-management-and-billing",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cost-management-and-billing.svg",
				Type: "Microsoft.Billing/billingAccounts",
			},
			{
				Name: "cost-management",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cost-management.svg",
				Type: "Microsoft.CostManagement/management",
			},
			{
				Name: "counter",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/counter.svg",
				Type: "Microsoft.DevOps/counters",
			},
			{
				Name: "cubes",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/cubes.svg",
				Type: "Microsoft.AnalysisServices/cubes",
			},
			{
				Name: "dashboard",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/dashboard.svg",
				Type: "Microsoft.Portal/dashboards",
			},
			{
				Name: "dev-console",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/dev-console.svg",
				Type: "Microsoft.DevOps/console",
			},
			{
				Name: "download",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/download.svg",
				Type: "Microsoft.Storage/downloads",
			},
			{
				Name: "error",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/error.svg",
				Type: "Microsoft.DevOps/errors",
			},
			{
				Name: "extensions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/extensions.svg",
				Type: "Microsoft.DevOps/extensions",
			},
			{
				Name: "feature-previews",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/feature-previews.svg",
				Type: "Microsoft.DevOps/features",
			},
			{
				Name: "file",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/file.svg",
				Type: "Microsoft.Storage/files",
			},
			{
				Name: "files",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/files.svg",
				Type: "Microsoft.Storage/files",
			},
			{
				Name: "folder-blank",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/folder-blank.svg",
				Type: "Microsoft.Storage/folders",
			},
			{
				Name: "folder-website",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/folder-website.svg",
				Type: "Microsoft.Web/folders",
			},
			{
				Name: "free-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/free-services.svg",
				Type: "Microsoft.Resources/freeServices",
			},
			{
				Name: "ftp",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/ftp.svg",
				Type: "Microsoft.Web/ftp",
			},
			{
				Name: "gear",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/gear.svg",
				Type: "Microsoft.DevOps/gear",
			},
			{
				Name: "globe-error",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/globe-error.svg",
				Type: "Microsoft.DevOps/globeErrors",
			},
			{
				Name: "globe-success",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/globe-success.svg",
				Type: "Microsoft.DevOps/globeSuccess",
			},
			{
				Name: "globe-warning",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/globe-warning.svg",
				Type: "Microsoft.DevOps/globeWarnings",
			},
			{
				Name: "guide",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/guide.svg",
				Type: "Microsoft.DevOps/guides",
			},
			{
				Name: "heart",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/heart.svg",
				Type: "Microsoft.DevOps/hearts",
			},
			{
				Name: "help-and-support",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/help-and-support.svg",
				Type: "Microsoft.Support/help",
			},
			{
				Name: "image",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/image.svg",
				Type: "Microsoft.Compute/images",
			},
			{
				Name: "information",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/information.svg",
				Type: "Microsoft.DevOps/information",
			},
			{
				Name: "input-output",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/input-output.svg",
				Type: "Microsoft.DevOps/io",
			},
			{
				Name: "journey-hub",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/journey-hub.svg",
				Type: "Microsoft.DevOps/journeyHub",
			},
			{
				Name: "launch-portal",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/launch-portal.svg",
				Type: "Microsoft.Portal/launch",
			},
			{
				Name: "learn",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/learn.svg",
				Type: "Microsoft.Learn/learningPaths",
			},
			{
				Name: "load-test",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/load-test.svg",
				Type: "Microsoft.LoadTest/loadTests",
			},
			{
				Name: "location",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/location.svg",
				Type: "Microsoft.DevOps/locations",
			},
			{
				Name: "log-streaming",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/log-streaming.svg",
				Type: "Microsoft.Insights/logProfiles",
			},
			{
				Name: "management-portal",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/management-portal.svg",
				Type: "Microsoft.Portal/management",
			},
			{
				Name: "marketplace-management",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/marketplace-management.svg",
				Type: "Microsoft.Marketplace/management",
			},
			{
				Name: "marketplace",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/marketplace.svg",
				Type: "Microsoft.Marketplace/offers",
			},
			{
				Name: "media-file",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/media-file.svg",
				Type: "Microsoft.Media/mediaFiles",
			},
			{
				Name: "media",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/media.svg",
				Type: "Microsoft.Media/mediaServices",
			},
			{
				Name: "mobile-engagement",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/mobile-engagement.svg",
				Type: "Microsoft.MobileEngagement/applications",
			},
			{
				Name: "mobile",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/mobile.svg",
				Type: "Microsoft.MobileEngagement/applications",
			},
			{
				Name: "module",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/module.svg",
				Type: "Microsoft.DevOps/modules",
			},
			{
				Name: "power-up",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/power-up.svg",
				Type: "Microsoft.PowerPlatform/powerUps",
			},
			{
				Name: "power",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/power.svg",
				Type: "Microsoft.PowerPlatform/power",
			},
			{
				Name: "powershell",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/powershell.svg",
				Type: "Microsoft.Automation/automationAccounts",
			},
			{
				Name: "preview-features",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/preview-features.svg",
				Type: "Microsoft.DevOps/features",
			},
			{
				Name: "process-explorer",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/process-explorer.svg",
				Type: "Microsoft.DevOps/processes",
			},
			{
				Name: "production-ready-database",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/production-ready-database.svg",
				Type: "Microsoft.Sql/servers/databases",
			},
			{
				Name: "quickstart-center",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/quickstart-center.svg",
				Type: "Microsoft.Portal/quickstart",
			},
			{
				Name: "recent",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/recent.svg",
				Type: "Microsoft.DevOps/recent",
			},
			{
				Name: "region-management",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/region-management.svg",
				Type: "Microsoft.DevOps/regions",
			},
			{
				Name: "reservations",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/reservations.svg",
				Type: "Microsoft.Capacity/reservationOrders",
			},
			{
				Name: "resource-explorer",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/resource-explorer.svg",
				Type: "Microsoft.Resources/resources",
			},
			{
				Name: "resource-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/resource-groups.svg",
				Type: "Microsoft.Resources/resourceGroups",
			},
			{
				Name: "resource-linked",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/resource-linked.svg",
				Type: "Microsoft.Resources/resources",
			},
			{
				Name: "scheduler",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/scheduler.svg",
				Type: "Microsoft.Scheduler/jobCollections",
			},
			{
				Name: "search-grid",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/search-grid.svg",
				Type: "Microsoft.Search/searchServices",
			},
			{
				Name: "search",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/search.svg",
				Type: "Microsoft.Search/searchServices",
			},
			{
				Name: "server-farm",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/server-farm.svg",
				Type: "Microsoft.Web/serverfarms",
			},
			{
				Name: "service-health",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/service-health.svg",
				Type: "Microsoft.ResourceHealth/availabilityStatuses",
			},
			{
				Name: "ssd",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/ssd.svg",
				Type: "Microsoft.Compute/disks",
			},
			{
				Name: "storage-azure-files",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/storage-azure-files.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "storage-container",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/storage-container.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "storage-queue",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/storage-queue.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "subscriptions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/subscriptions.svg",
				Type: "Microsoft.Resources/subscriptions",
			},
			{
				Name: "table",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/table.svg",
				Type: "Microsoft.Storage/storageAccounts",
			},
			{
				Name: "tag",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/tag.svg",
				Type: "Microsoft.Resources/tags",
			},
			{
				Name: "tags",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/tags.svg",
				Type: "Microsoft.Resources/tags",
			},
			{
				Name: "templates",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/templates.svg",
				Type: "Microsoft.Resources/deployments",
			},
			{
				Name: "tfs-vc-repository",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/tfs-vc-repository.svg",
				Type: "Microsoft.DevOps/repositories",
			},
			{
				Name: "toolbox",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/toolbox.svg",
				Type: "Microsoft.DevOps/toolboxes",
			},
		},
	},
	{
		Category:     "Intune",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/intune.svg",
		Items: []AzureNodeItem{
			{
				Name: "client-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/client-apps.svg",
				Type: "Microsoft.Intune/clientApps",
			},
			{
				Name: "device-compliance",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/device-compliance.svg",
				Type: "Microsoft.Intune/deviceCompliance",
			},
			{
				Name: "device-configuration",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/device-configuration.svg",
				Type: "Microsoft.Intune/deviceConfiguration",
			},
			{
				Name: "device-enrollment",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/device-enrollment.svg",
				Type: "Microsoft.Intune/deviceEnrollment",
			},
			{
				Name: "device-security-apple",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/device-security-apple.svg",
				Type: "Microsoft.Intune/deviceSecurity",
			},
			{
				Name: "device-security-google",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/device-security-google.svg",
				Type: "Microsoft.Intune/deviceSecurity",
			},
			{
				Name: "device-security-windows",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/device-security-windows.svg",
				Type: "Microsoft.Intune/deviceSecurity",
			},
			{
				Name: "devices",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/devices.svg",
				Type: "Microsoft.Intune/devices",
			},
			{
				Name: "ebooks",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/ebooks.svg",
				Type: "Microsoft.Intune/ebooks",
			},
			{
				Name: "entra-identity-roles-and-administrators",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/entra-identity-roles-and-administrators.svg",
				Type: "Microsoft.Entra/roles",
			},
			{
				Name: "exchange-access",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/exchange-access.svg",
				Type: "Microsoft.Intune/exchangeAccess",
			},
			{
				Name: "intune-app-protection",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/intune-app-protection.svg",
				Type: "Microsoft.Intune/appProtection",
			},
			{
				Name: "intune-for-education",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/intune-for-education.svg",
				Type: "Microsoft.Intune/education",
			},
			{
				Name: "intune",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/intune.svg",
				Type: "Microsoft.Intune/intune",
			},
			{
				Name: "mindaro",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/mindaro.svg",
				Type: "Microsoft.Mindaro/services",
			},
			{
				Name: "security-baselines",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/security-baselines.svg",
				Type: "Microsoft.Intune/securityBaselines",
			},
			{
				Name: "software-updates",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/software-updates.svg",
				Type: "Microsoft.Intune/softwareUpdates",
			},
			{
				Name: "tenant-status",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/intune/tenant-status.svg",
				Type: "Microsoft.Intune/tenantStatus",
			},
		},
	},
	{
		Category:     "Management and Governance",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/management-groups.svg",
		Items: []AzureNodeItem{
			{
				Name: "management-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/management-groups.svg",
				Type: "Microsoft.Management/managementGroups",
			},
			{
				Name: "activity-log",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/activity-log.svg",
				Type: "Microsoft.Insights/eventTypes",
			},
			{
				Name: "advisor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/advisor.svg",
				Type: "Microsoft.Advisor/recommendations",
			},
			{
				Name: "alerts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/alerts.svg",
				Type: "Microsoft.Insights/alerts",
			},
			{
				Name: "arc-machines",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/arc-machines.svg",
				Type: "Microsoft.HybridCompute/machines",
			},
			{
				Name: "automation-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/automation-accounts.svg",
				Type: "Microsoft.Automation/automationAccounts",
			},
			{
				Name: "azure-arc",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/azure-arc.svg",
				Type: "Microsoft.HybridCompute/machines",
			},
			{
				Name: "azure-lighthouse",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/azure-lighthouse.svg",
				Type: "Microsoft.ManagedServices/registrationDefinitions",
			},
			{
				Name: "blueprints",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/blueprints.svg",
				Type: "Microsoft.Blueprint/blueprints",
			},
			{
				Name: "compliance",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/compliance.svg",
				Type: "Microsoft.PolicyInsights/complianceResults",
			},
			{
				Name: "cost-management-and-billing",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/cost-management-and-billing.svg",
				Type: "Microsoft.Billing/billingAccounts",
			},
			{
				Name: "customer-lockbox-for-microsoft-azure",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/customer-lockbox-for-microsoft-azure.svg",
				Type: "Microsoft.CustomerLockbox/requests",
			},
			{
				Name: "diagnostics-settings",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/diagnostics-settings.svg",
				Type: "Microsoft.Insights/diagnosticSettings",
			},
			{
				Name: "education",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/education.svg",
				Type: "Microsoft.Education/schools",
			},
			{
				Name: "intune-trends",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/intune-trends.svg",
				Type: "Microsoft.Intune/trends",
			},
			{
				Name: "log-analytics-workspaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/log-analytics-workspaces.svg",
				Type: "Microsoft.OperationalInsights/workspaces",
			},
			{
				Name: "machinesazurearc",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/machinesazurearc.svg",
				Type: "Microsoft.HybridCompute/machines",
			},
			{
				Name: "managed-applications-center",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/managed-applications-center.svg",
				Type: "Microsoft.Solutions/applications",
			},
			{
				Name: "managed-desktop",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/managed-desktop.svg",
				Type: "Microsoft.ManagedDesktop/managedDesktops",
			},
			{
				Name: "metrics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/metrics.svg",
				Type: "Microsoft.Insights/metrics",
			},
			{
				Name: "monitor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/monitor.svg",
				Type: "Microsoft.Insights/monitors",
			},
			{
				Name: "my-customers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/my-customers.svg",
				Type: "Microsoft.CustomerInsights/hubs",
			},
			{
				Name: "policy",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/policy.svg",
				Type: "Microsoft.Authorization/policyDefinitions",
			},
			{
				Name: "recovery-services-vaults",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/recovery-services-vaults.svg",
				Type: "Microsoft.RecoveryServices/vaults",
			},
			{
				Name: "resource-graph-explorer",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/resource-graph-explorer.svg",
				Type: "Microsoft.ResourceGraph/resources",
			},
			{
				Name: "resources-provider",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/resources-provider.svg",
				Type: "Microsoft.Resources/resources",
			},
			{
				Name: "scheduler-job-collections",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/scheduler-job-collections.svg",
				Type: "Microsoft.Scheduler/jobCollections",
			},
			{
				Name: "service-catalog-mad",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/service-catalog-mad.svg",
				Type: "Microsoft.ServiceCatalog/serviceCatalogs",
			},
			{
				Name: "service-providers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/service-providers.svg",
				Type: "Microsoft.ServiceProvider/serviceProviders",
			},
			{
				Name: "solutions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/solutions.svg",
				Type: "Microsoft.Solutions/applications",
			},
			{
				Name: "universal-print",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/universal-print.svg",
				Type: "Microsoft.UniversalPrint/printers",
			},
			{
				Name: "user-privacy",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/management_and_governance/user-privacy.svg",
				Type: "Microsoft.Privacy/privacySettings",
			},
		},
	},
	{
		Category:     "Migration",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/azure-migrate.svg",
		Items: []AzureNodeItem{
			{
				Name: "azure-database-migration-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/azure-database-migration-services.svg",
				Type: "Microsoft.DataMigration/services",
			},
			{
				Name: "azure-databox-gateway",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/azure-databox-gateway.svg",
				Type: "Microsoft.DataBoxEdge/dataBoxEdgeDevices",
			},
			{
				Name: "azure-migrate",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/azure-migrate.svg",
				Type: "Microsoft.Migrate/migrateProjects",
			},
			{
				Name: "azure-stack-edge",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/azure-stack-edge.svg",
				Type: "Microsoft.DataBoxEdge/dataBoxEdgeDevices",
			},
			{
				Name: "cost-management-and-billing",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/cost-management-and-billing.svg",
				Type: "Microsoft.Billing/billingAccounts",
			},
			{
				Name: "data-box",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/data-box.svg",
				Type: "Microsoft.DataBox/jobs",
			},
			{
				Name: "recovery-services-vaults",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/migration/recovery-services-vaults.svg",
				Type: "Microsoft.RecoveryServices/vaults",
			},
		},
	},
	{
		Category:     "Mobile",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/mobile/app-services.svg",
		Items: []AzureNodeItem{
			{
				Name: "app-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/mobile/app-services.svg",
				Type: "Microsoft.Web/sites",
			},
			{
				Name: "notification-hubs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/mobile/notification-hubs.svg",
				Type: "Microsoft.NotificationHubs/namespaces",
			},
			{
				Name: "power-platform",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/mobile/power-platform.svg",
				Type: "Microsoft.PowerPlatform/enterprisePolicies",
			},
		},
	},
	{
		Category:     "Monitor",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/log-analytics-workspaces.svg",
		Items: []AzureNodeItem{
			{
				Name: "activity-log",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/activity-log.svg",
				Type: "Microsoft.Insights/eventTypes",
			},
			{
				Name: "application-insights",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/application-insights.svg",
				Type: "microsoft.insights/components",
			},
			{
				Name: "auto-scale",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/auto-scale.svg",
				Type: "Microsoft.Insights/autoscaleSettings",
			},
			{
				Name: "azure-monitors-for-sap-solutions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/azure-monitors-for-sap-solutions.svg",
				Type: "Microsoft.AzureMonitorForSAPSolutions/monitors",
			},
			{
				Name: "azure-workbooks",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/azure-workbooks.svg",
				Type: "Microsoft.Insights/workbooks",
			},
			{
				Name: "change-analysis",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/change-analysis.svg",
				Type: "Microsoft.ChangeAnalysis/changeAnalysis",
			},
			{
				Name: "diagnostics-settings",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/diagnostics-settings.svg",
				Type: "Microsoft.Insights/diagnosticSettings",
			},
			{
				Name: "log-analytics-workspaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/log-analytics-workspaces.svg",
				Type: "Microsoft.OperationalInsights/workspaces",
			},
			{
				Name: "metrics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/metrics.svg",
				Type: "Microsoft.Insights/metrics",
			},
			{
				Name: "monitor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/monitor.svg",
				Type: "Microsoft.Insights/monitors",
			},
			{
				Name: "network-watcher",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/monitor/network-watcher.svg",
				Type: "Microsoft.Network/networkWatchers",
			},
		},
	},
	{
		Category:     "Other",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/general/backlog.svg",
		Items: []AzureNodeItem{
			{
				Name: "aks-automatic",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/aks-automatic.svg",
				Type: "Microsoft.ContainerService/managedClusters",
			},
			{
				Name: "aks-istio",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/aks-istio.svg",
				Type: "Microsoft.ContainerService/managedClusters",
			},
			{
				Name: "app-compliance-automation",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/app-compliance-automation.svg",
				Type: "Microsoft.AppComplianceAutomation/automation",
			},
			{
				Name: "app-registrations",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/app-registrations.svg",
				Type: "Microsoft.AppRegistrations/registrations",
			},
			{
				Name: "app-space-component",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/app-space-component.svg",
				Type: "Microsoft.AppSpace/components",
			},
			{
				Name: "aquila",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/aquila.svg",
				Type: "Microsoft.Aquila/services",
			},
			{
				Name: "arc-data-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/arc-data-services.svg",
				Type: "Microsoft.AzureArcData/dataControllers",
			},
			{
				Name: "arc-kubernetes",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/arc-kubernetes.svg",
				Type: "Microsoft.Kubernetes/connectedClusters",
			},
			{
				Name: "arc-sql-managed-instance",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/arc-sql-managed-instance.svg",
				Type: "Microsoft.AzureArcData/sqlManagedInstances",
			},
			{
				Name: "arc-sql-server",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/arc-sql-server.svg",
				Type: "Microsoft.AzureArcData/sqlServers",
			},
			{
				Name: "defender-cm-local-manager",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-cm-local-manager.svg",
				Type: "Microsoft.Defender/cmLocalManagers",
			},
			{
				Name: "defender-dcs-controller",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-dcs-controller.svg",
				Type: "Microsoft.Defender/dcsControllers",
			},
			{
				Name: "defender-distributer-control-system",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-distributer-control-system.svg",
				Type: "Microsoft.Defender/distributerControlSystems",
			},
			{
				Name: "defender-engineering-station",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-engineering-station.svg",
				Type: "Microsoft.Defender/engineeringStations",
			},
			{
				Name: "defender-external-management",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-external-management.svg",
				Type: "Microsoft.Defender/externalManagement",
			},
			{
				Name: "defender-freezer-monitor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-freezer-monitor.svg",
				Type: "Microsoft.Defender/freezerMonitors",
			},
			{
				Name: "defender-historian",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-historian.svg",
				Type: "Microsoft.Defender/historians",
			},
			{
				Name: "defender-hmi",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-hmi.svg",
				Type: "Microsoft.Defender/hmis",
			},
			{
				Name: "defender-industrial-packaging-system",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-industrial-packaging-system.svg",
				Type: "Microsoft.Defender/industrialPackagingSystems",
			},
			{
				Name: "defender-industrial-printer",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-industrial-printer.svg",
				Type: "Microsoft.Defender/industrialPrinters",
			},
			{
				Name: "defender-industrial-robot",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-industrial-robot.svg",
				Type: "Microsoft.Defender/industrialRobots",
			},
			{
				Name: "defender-industrial-scale-system",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-industrial-scale-system.svg",
				Type: "Microsoft.Defender/industrialScaleSystems",
			},
			{
				Name: "defender-marquee",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-marquee.svg",
				Type: "Microsoft.Defender/marquees",
			},
			{
				Name: "defender-meter",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-meter.svg",
				Type: "Microsoft.Defender/meters",
			},
			{
				Name: "defender-plc",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-plc.svg",
				Type: "Microsoft.Defender/plcs",
			},
			{
				Name: "defender-pneumatic-device",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-pneumatic-device.svg",
				Type: "Microsoft.Defender/pneumaticDevices",
			},
			{
				Name: "defender-programable-board",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-programable-board.svg",
				Type: "Microsoft.Defender/programableBoards",
			},
			{
				Name: "defender-relay",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-relay.svg",
				Type: "Microsoft.Defender/relays",
			},
			{
				Name: "defender-robot-controller",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-robot-controller.svg",
				Type: "Microsoft.Defender/robotControllers",
			},
			{
				Name: "defender-rtu",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-rtu.svg",
				Type: "Microsoft.Defender/rtus",
			},
			{
				Name: "defender-sensor",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-sensor.svg",
				Type: "Microsoft.Defender/sensors",
			},
			{
				Name: "defender-slot",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-slot.svg",
				Type: "Microsoft.Defender/slots",
			},
			{
				Name: "defender-web-guiding-system",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/other/defender-web-guiding-system.svg",
				Type: "Microsoft.Defender/webGuidingSystems",
			},
		},
	},
	{
		Category:     "Security",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/azure-sentinel.svg",
		Items: []AzureNodeItem{
			{
				Name: "application-security-groups",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/application-security-groups.svg",
				Type: "Microsoft.Network/applicationSecurityGroups",
			},
			{
				Name: "azure-information-protection",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/azure-information-protection.svg",
				Type: "Microsoft.InformationProtection/policies",
			},
			{
				Name: "azure-sentinel",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/azure-sentinel.svg",
				Type: "Microsoft.OperationalInsights/workspaces",
			},
			{
				Name: "conditional-access",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/conditional-access.svg",
				Type: "Microsoft.AAD/conditionalAccessPolicies",
			},
			{
				Name: "detonation",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/detonation.svg",
				Type: "Microsoft.Detonation/detonationServices",
			},
			{
				Name: "entra-identity-risky-signins",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/entra-identity-risky-signins.svg",
				Type: "Microsoft.Entra/identityRiskySignins",
			},
			{
				Name: "entra-identity-risky-users",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/entra-identity-risky-users.svg",
				Type: "Microsoft.Entra/identityRiskyUsers",
			},
			{
				Name: "extendedsecurityupdates",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/extendedsecurityupdates.svg",
				Type: "Microsoft.ESU/updates",
			},
			{
				Name: "identity-secure-score",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/identity-secure-score.svg",
				Type: "Microsoft.Identity/secureScores",
			},
			{
				Name: "key-vaults",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/key-vaults.svg",
				Type: "Microsoft.KeyVault/vaults",
			},
			{
				Name: "microsoft-defender-easm",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/microsoft-defender-easm.svg",
				Type: "Microsoft.Defender/easm",
			},
			{
				Name: "microsoft-defender-for-cloud",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/microsoft-defender-for-cloud.svg",
				Type: "Microsoft.Security/defenderForCloud",
			},
			{
				Name: "microsoft-defender-for-iot",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/microsoft-defender-for-iot.svg",
				Type: "Microsoft.IoTDefender/defenderForIoT",
			},
			{
				Name: "multifactor-authentication",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/multifactor-authentication.svg",
				Type: "Microsoft.AAD/multiFactorAuth",
			},
			{
				Name: "user-settings",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/security/user-settings.svg",
				Type: "Microsoft.AAD/userSettings",
			},
		},
	},
	{
		Category:     "Azure Ecosystem",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_ecosystem/applens.svg",
		Items: []AzureNodeItem{
			{
				Name: "applens",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_ecosystem/applens.svg",
				Type: "Microsoft.AppLens/lenses",
			},
			{
				Name: "azure-hybrid-center",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_ecosystem/azure-hybrid-center.svg",
				Type: "Microsoft.HybridCenter/centers",
			},
			{
				Name: "collaborative-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_ecosystem/collaborative-service.svg",
				Type: "Microsoft.Collaborative/services",
			},
		},
	},
	{
		Category:     "IoT",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/iot-hub.svg",
		Items: []AzureNodeItem{
			{
				Name: "azure-cosmos-db",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/azure-cosmos-db.svg",
				Type: "Microsoft.DocumentDB/databaseAccounts",
			},
			{
				Name: "azure-databox-gateway",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/azure-databox-gateway.svg",
				Type: "Microsoft.DataBox/gateways",
			},
			{
				Name: "azure-iot-operations",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/azure-iot-operations.svg",
				Type: "Microsoft.IoTOperations/operations",
			},
			{
				Name: "azure-maps-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/azure-maps-accounts.svg",
				Type: "Microsoft.Maps/accounts",
			},
			{
				Name: "azure-stack-hci-sizer",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/azure-stack-hci-sizer.svg",
				Type: "Microsoft.HCI/sizers",
			},
			{
				Name: "azure-stack",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/azure-stack.svg",
				Type: "Microsoft.AzureStack/azureStacks",
			},
			{
				Name: "device-provisioning-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/device-provisioning-services.svg",
				Type: "Microsoft.Devices/provisioningServices",
			},
			{
				Name: "digital-twins",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/digital-twins.svg",
				Type: "Microsoft.DigitalTwins/digitalTwinsInstances",
			},
			{
				Name: "event-grid-subscriptions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/event-grid-subscriptions.svg",
				Type: "Microsoft.EventGrid/eventSubscriptions",
			},
			{
				Name: "event-hub-clusters",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/event-hub-clusters.svg",
				Type: "Microsoft.EventHub/clusters",
			},
			{
				Name: "event-hubs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/event-hubs.svg",
				Type: "Microsoft.EventHub/namespaces",
			},
			{
				Name: "industrial-iot",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/industrial-iot.svg",
				Type: "Microsoft.IndustrialIoT/industrialIoTServices",
			},
			{
				Name: "iot-central-applications",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/iot-central-applications.svg",
				Type: "Microsoft.IoTCentral/applications",
			},
			{
				Name: "iot-edge",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/iot-edge.svg",
				Type: "Microsoft.Devices/IotHubs",
			},
			{
				Name: "iot-hub",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/iot-hub.svg",
				Type: "Microsoft.Devices/IotHubs",
			},
			{
				Name: "logic-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/logic-apps.svg",
				Type: "Microsoft.Logic/workflows",
			},
			{
				Name: "machine-learning-studio-(classic)-web-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/machine-learning-studio-(classic)-web-services.svg",
				Type: "Microsoft.MachineLearning/webServices",
			},
			{
				Name: "machine-learning-studio-web-service-plans",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/machine-learning-studio-web-service-plans.svg",
				Type: "Microsoft.MachineLearning/webServicePlans",
			},
			{
				Name: "machine-learning-studio-workspaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/machine-learning-studio-workspaces.svg",
				Type: "Microsoft.MachineLearning/workspaces",
			},
			{
				Name: "notification-hub-namespaces",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/notification-hub-namespaces.svg",
				Type: "Microsoft.NotificationHubs/namespaces",
			},
			{
				Name: "notification-hubs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/notification-hubs.svg",
				Type: "Microsoft.NotificationHubs/namespaces",
			},
			{
				Name: "stack-hci-premium",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/stack-hci-premium.svg",
				Type: "Microsoft.HCI/premium",
			},
			{
				Name: "stream-analytics-jobs",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/stream-analytics-jobs.svg",
				Type: "Microsoft.StreamAnalytics/streamingJobs",
			},
			{
				Name: "time-series-data-sets",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/time-series-data-sets.svg",
				Type: "Microsoft.TimeSeriesInsights/environments",
			},
			{
				Name: "time-series-insights-access-policies",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/time-series-insights-access-policies.svg",
				Type: "Microsoft.TimeSeriesInsights/accessPolicies",
			},
			{
				Name: "time-series-insights-environments",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/time-series-insights-environments.svg",
				Type: "Microsoft.TimeSeriesInsights/environments",
			},
			{
				Name: "time-series-insights-event-sources",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/time-series-insights-event-sources.svg",
				Type: "Microsoft.TimeSeriesInsights/eventSources",
			},
			{
				Name: "windows10-core-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/iot/windows10-core-services.svg",
				Type: "Microsoft.Windows10/coreServices",
			},
		},
	},
	{
		Category:     "Integration",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/api-connections.svg",
		Items: []AzureNodeItem{
			{
				Name: "api-connections",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/api-connections.svg",
				Type: "Microsoft.Web/connections",
			},
			{
				Name: "api-management-services",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/api-management-services.svg",
				Type: "Microsoft.ApiManagement/service",
			},
			{
				Name: "app-configuration",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/app-configuration.svg",
				Type: "Microsoft.AppConfiguration/configurationStores",
			},
			{
				Name: "azure-api-for-fhir",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/azure-api-for-fhir.svg",
				Type: "Microsoft.HealthcareApis/services",
			},
			{
				Name: "azure-data-catalog",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/azure-data-catalog.svg",
				Type: "Microsoft.DataCatalog/catalogs",
			},
			{
				Name: "azure-databox-gateway",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/azure-databox-gateway.svg",
				Type: "Microsoft.DataBox/gateways",
			},
			{
				Name: "azure-service-bus",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/azure-service-bus.svg",
				Type: "Microsoft.ServiceBus/namespaces",
			},
			{
				Name: "azure-sql-server-stretch-databases",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/azure-sql-server-stretch-databases.svg",
				Type: "Microsoft.Sql/servers",
			},
			{
				Name: "azure-stack-edge",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/azure-stack-edge.svg",
				Type: "Microsoft.DataBoxEdge/dataBoxEdgeDevices",
			},
			{
				Name: "data-factories",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/data-factories.svg",
				Type: "Microsoft.DataFactory/factories",
			},
			{
				Name: "event-grid-domains",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/event-grid-domains.svg",
				Type: "Microsoft.EventGrid/domains",
			},
			{
				Name: "event-grid-subscriptions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/event-grid-subscriptions.svg",
				Type: "Microsoft.EventGrid/eventSubscriptions",
			},
			{
				Name: "event-grid-topics",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/event-grid-topics.svg",
				Type: "Microsoft.EventGrid/topics",
			},
			{
				Name: "integration-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/integration-accounts.svg",
				Type: "Microsoft.Logic/integrationAccounts",
			},
			{
				Name: "integration-environments",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/integration-environments.svg",
				Type: "Microsoft.Integration/environments",
			},
			{
				Name: "integration-service-environments",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/integration-service-environments.svg",
				Type: "Microsoft.Integration/serviceEnvironments",
			},
			{
				Name: "logic-apps-custom-connector",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/logic-apps-custom-connector.svg",
				Type: "Microsoft.Logic/customConnectors",
			},
			{
				Name: "logic-apps",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/logic-apps.svg",
				Type: "Microsoft.Logic/workflows",
			},
			{
				Name: "partner-namespace",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/partner-namespace.svg",
				Type: "Microsoft.EventGrid/partnerNamespaces",
			},
			{
				Name: "partner-registration",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/partner-registration.svg",
				Type: "Microsoft.EventGrid/partnerRegistrations",
			},
			{
				Name: "partner-topic",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/partner-topic.svg",
				Type: "Microsoft.EventGrid/partnerTopics",
			},
			{
				Name: "power-platform",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/power-platform.svg",
				Type: "Microsoft.PowerPlatform/enterprisePolicies",
			},
			{
				Name: "relays",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/relays.svg",
				Type: "Microsoft.Relay/namespaces",
			},
			{
				Name: "sendgrid-accounts",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/sendgrid-accounts.svg",
				Type: "Microsoft.SendGrid/accounts",
			},
			{
				Name: "software-as-a-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/software-as-a-service.svg",
				Type: "Microsoft.SaaS/services",
			},
			{
				Name: "sql-data-warehouses",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/sql-data-warehouses.svg",
				Type: "Microsoft.Sql/servers",
			},
			{
				Name: "storsimple-device-managers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/storsimple-device-managers.svg",
				Type: "Microsoft.StorSimple/managers",
			},
			{
				Name: "system-topic",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/integration/system-topic.svg",
				Type: "Microsoft.EventGrid/systemTopics",
			},
		},
	},
	{
		Category:     "Azure Stack",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/plans.svg",
		Items: []AzureNodeItem{
			{
				Name: "capacity",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/capacity.svg",
				Type: "Microsoft.Capacity/reservationOrders",
			},
			{
				Name: "infrastructure-backup",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/infrastructure-backup.svg",
				Type: "Microsoft.AzureStack/infrastructureBackups",
			},
			{
				Name: "multi-tenancy",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/multi-tenancy.svg",
				Type: "Microsoft.AzureStack/multiTenancy",
			},
			{
				Name: "offers",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/offers.svg",
				Type: "Microsoft.AzureStack/offers",
			},
			{
				Name: "plans",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/plans.svg",
				Type: "Microsoft.AzureStack/plans",
			},
			{
				Name: "updates",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/updates.svg",
				Type: "Microsoft.AzureStack/updates",
			},
			{
				Name: "user-subscriptions",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/azure_stack/user-subscriptions.svg",
				Type: "Microsoft.AzureStack/userSubscriptions",
			},
		},
	},
	{
		Category:     "Blockchain",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/blockchain-applications.svg",
		Items: []AzureNodeItem{
			{
				Name: "abs-member",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/abs-member.svg",
				Type: "Microsoft.Blockchain/blockchainMembers",
			},
			{
				Name: "azure-blockchain-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/azure-blockchain-service.svg",
				Type: "Microsoft.Blockchain/blockchainServices",
			},
			{
				Name: "azure-token-service",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/azure-token-service.svg",
				Type: "Microsoft.Blockchain/tokenServices",
			},
			{
				Name: "blockchain-applications",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/blockchain-applications.svg",
				Type: "Microsoft.Blockchain/blockchainApplications",
			},
			{
				Name: "consortium",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/consortium.svg",
				Type: "Microsoft.Blockchain/consortiums",
			},
			{
				Name: "outbound-connection",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/blockchain/outbound-connection.svg",
				Type: "Microsoft.Blockchain/outboundConnections",
			},
		},
	},
	{
		Category:     "Hybrid and Multicloud",
		CategoryIcon: "https://spnodedata.blob.core.windows.net/nodes/azure_clean/hybrid_and_multicloud/azure-operator-5g-core.svg",
		Items: []AzureNodeItem{
			{
				Name: "azure-operator-5g-core",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/hybrid_and_multicloud/azure-operator-5g-core.svg",
				Type: "Microsoft.Operator/5gCores",
			},
			{
				Name: "azure-operator-insights",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/hybrid_and_multicloud/azure-operator-insights.svg",
				Type: "Microsoft.Operator/insights",
			},
			{
				Name: "azure-operator-nexus",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/hybrid_and_multicloud/azure-operator-nexus.svg",
				Type: "Microsoft.Operator/nexus",
			},
			{
				Name: "azure-operator-service-manager",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/hybrid_and_multicloud/azure-operator-service-manager.svg",
				Type: "Microsoft.Operator/serviceManagers",
			},
			{
				Name: "azure-programmable-connectivity",
				URL:  "https://spnodedata.blob.core.windows.net/nodes/azure_clean/hybrid_and_multicloud/azure-programmable-connectivity.svg",
				Type: "Microsoft.Operator/programmableConnectivity",
			},
		},
	},
}
