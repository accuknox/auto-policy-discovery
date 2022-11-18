package libs

import (
	"database/sql"
	"errors"
	"sync"

	cfg "github.com/accuknox/auto-policy-discovery/src/config"
	logger "github.com/accuknox/auto-policy-discovery/src/logging"
	"github.com/accuknox/auto-policy-discovery/src/types"
	"github.com/robfig/cron"
)

// ================= //
// == Network Log == //
// ================= //

// LastFlowID network flow between [ startTime <= time < endTime ]
var LastFlowID int64 = 0

// ==================== //
// == Network Policy == //
// ==================== //

func GetNetworkPolicies(cfg types.ConfigDB, cluster, namespace, status, nwtype, rule string) []types.KnoxNetworkPolicy {
	results := []types.KnoxNetworkPolicy{}

	if cfg.DBDriver == "mysql" {
		docs, err := GetNetworkPoliciesFromMySQL(cfg, cluster, namespace, status, nwtype, rule)
		if err != nil {
			return results
		}
		results = docs
	} else if cfg.DBDriver == "sqlite3" {
		docs, err := GetNetworkPoliciesFromSQLite(cfg, cluster, namespace, status)
		if err != nil {
			return results
		}
		results = docs
	}

	return results
}

func GetNetworkPoliciesBySelector(cfg types.ConfigDB, cluster, namespace, status string, selector map[string]string) ([]types.KnoxNetworkPolicy, error) {
	results := []types.KnoxNetworkPolicy{}

	if cfg.DBDriver == "mysql" {
		docs, err := GetNetworkPoliciesFromMySQL(cfg, cluster, namespace, status, "", "")
		if err != nil {
			return nil, err
		}
		results = docs
	} else if cfg.DBDriver == "sqlite3" {
		docs, err := GetNetworkPoliciesFromSQLite(cfg, cluster, namespace, status)
		if err != nil {
			return nil, err
		}
		results = docs
	} else {
		return results, nil
	}

	filtered := []types.KnoxNetworkPolicy{}
	for _, policy := range results {
		matched := true
		for k, v := range selector {
			val := policy.Spec.Selector.MatchLabels[k]
			if val != v {
				matched = false
				break
			}
		}

		if matched {
			filtered = append(filtered, policy)
		}
	}

	return filtered, nil
}

func UpdateOutdatedNetworkPolicy(cfg types.ConfigDB, outdatedPolicy string, latestPolicy string) {
	if cfg.DBDriver == "mysql" {
		if err := UpdateOutdatedNetworkPolicyFromMySQL(cfg, outdatedPolicy, latestPolicy); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := UpdateOutdatedNetworkPolicyFromSQLite(cfg, outdatedPolicy, latestPolicy); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func UpdateNetworkPolicies(cfg types.ConfigDB, policies []types.KnoxNetworkPolicy) {
	for _, policy := range policies {
		UpdateNetworkPolicy(cfg, policy)
	}
}

func UpdateNetworkPolicy(cfg types.ConfigDB, policy types.KnoxNetworkPolicy) {
	if cfg.DBDriver == "mysql" {
		if err := UpdateNetworkPolicyToMySQL(cfg, policy); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := UpdateNetworkPolicyToSQLite(cfg, policy); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func InsertNetworkPolicies(cfg types.ConfigDB, policies []types.KnoxNetworkPolicy) {
	if cfg.DBDriver == "mysql" {
		if err := InsertNetworkPoliciesToMySQL(cfg, policies); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := InsertNetworkPoliciesToSQLite(cfg, policies); err != nil {
			log.Error().Msg(err.Error())
		}
	}

}

// ================ //
// == System Log == //
// ================ //

// LastSyslogID system log between [ startTime <= time < endTime ]
var LastSyslogID int64 = 0

// ================== //
// == System Alert == //
// ================== //

// LastSysAlertID system_alert between [ startTime <= time < endTime ]
var LastSysAlertID int64 = 0

// =================== //
// == System Policy == //
// =================== //

func UpdateOutdatedSystemPolicy(cfg types.ConfigDB, outdatedPolicy string, latestPolicy string) {
	if cfg.DBDriver == "mysql" {
		if err := UpdateOutdatedNetworkPolicyFromMySQL(cfg, outdatedPolicy, latestPolicy); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := UpdateOutdatedNetworkPolicyFromSQLite(cfg, outdatedPolicy, latestPolicy); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func GetSystemPolicies(cfg types.ConfigDB, namespace, status string) []types.KnoxSystemPolicy {
	results := []types.KnoxSystemPolicy{}

	if cfg.DBDriver == "mysql" {
		docs, err := GetSystemPoliciesFromMySQL(cfg, namespace, status)
		if err != nil {
			return results
		}
		results = docs
	} else if cfg.DBDriver == "sqlite3" {
		docs, err := GetSystemPoliciesFromSQLite(cfg, namespace, status)
		if err != nil {
			return results
		}
		results = docs
	}

	return results
}

func InsertSystemPolicies(cfg types.ConfigDB, policies []types.KnoxSystemPolicy) {
	if cfg.DBDriver == "mysql" {
		if err := InsertSystemPoliciesToMySQL(cfg, policies); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := InsertSystemPoliciesToSQLite(cfg, policies); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func UpdateSystemPolicy(cfg types.ConfigDB, policy types.KnoxSystemPolicy) {
	if cfg.DBDriver == "mysql" {
		if err := UpdateSystemPolicyToMySQL(cfg, policy); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := UpdateSystemPolicyToSQLite(cfg, policy); err != nil {
			log.Error().Msg(err.Error())
		}
	}

}

func GetWorkloadProcessFileSet(cfg types.ConfigDB, wpfs types.WorkloadProcessFileSet) (map[types.WorkloadProcessFileSet][]string, types.PolicyNameMap, error) {
	if cfg.DBDriver == "mysql" {
		res, pnMap, err := GetWorkloadProcessFileSetMySQL(cfg, wpfs)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		return res, pnMap, err
	} else if cfg.DBDriver == "sqlite3" {
		res, pnMap, err := GetWorkloadProcessFileSetSQLite(cfg, wpfs)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		return res, pnMap, err
	}
	return nil, nil, errors.New("no db driver")
}

func InsertWorkloadProcessFileSet(cfg types.ConfigDB, wpfs types.WorkloadProcessFileSet, fs []string) error {
	if cfg.DBDriver == "mysql" {
		return InsertWorkloadProcessFileSetMySQL(cfg, wpfs, fs)
	} else if cfg.DBDriver == "sqlite3" {
		return InsertWorkloadProcessFileSetSQLite(cfg, wpfs, fs)
	}
	return errors.New("no db driver")
}

func ClearWPFSDb(cfg types.ConfigDB, wpfs types.WorkloadProcessFileSet, duration int64) error {
	if cfg.DBDriver == "mysql" {
		return ClearWPFSDbMySQL(cfg, wpfs, duration)
	} else if cfg.DBDriver == "sqlite3" {
		return ClearWPFSDbSQLite(cfg, wpfs, duration)
	}
	return errors.New("no db driver")
}

// =========== //
// == Table == //
// =========== //

func ClearDBTables(cfg types.ConfigDB) {
	if cfg.DBDriver == "mysql" {
		if err := ClearDBTablesMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := ClearDBTablesSQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func ClearNetworkDBTable(cfg types.ConfigDB) {
	if cfg.DBDriver == "mysql" {
		if err := ClearNetworkDBTableMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func CreateTablesIfNotExist(cfg types.ConfigDB) {
	if cfg.DBDriver == "mysql" {
		if err := CreateTableNetworkPolicyMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemPolicyMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableWorkLoadProcessFileSetMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemLogsMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableNetworkLogsMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreatePolicyTableMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	} else if cfg.DBDriver == "sqlite3" {
		if err := CreateTableNetworkPolicySQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemPolicySQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableWorkLoadProcessFileSetSQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemLogsSQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableNetworkLogsSQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreatePolicyTableSQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateSystemSummaryTableSQLite(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

// =================== //
// == Observability == //
// =================== //
func UpdateOrInsertKubearmorLogs(cfg types.ConfigDB, kubearmorLogMap map[types.KubeArmorLog]int) error {
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		err = UpdateOrInsertKubearmorLogsMySQL(cfg, kubearmorLogMap)
	} else if cfg.DBDriver == "sqlite3" {
		err = UpdateOrInsertKubearmorLogsSQLite(cfg, kubearmorLogMap)
	}
	return err
}

func GetKubearmorLogs(cfg types.ConfigDB, filterLog types.KubeArmorLog) ([]types.KubeArmorLog, []uint32, error) {
	kubearmorLog := []types.KubeArmorLog{}
	totalCount := []uint32{}
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		kubearmorLog, totalCount, err = GetSystemLogsMySQL(cfg, filterLog)
	} else if cfg.DBDriver == "sqlite3" {
		kubearmorLog, totalCount, err = GetSystemLogsSQLite(cfg, filterLog)
	}
	return kubearmorLog, totalCount, err
}

func UpdateOrInsertCiliumLogs(cfg types.ConfigDB, ciliumLogs []types.CiliumLog) error {
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		err = UpdateOrInsertCiliumLogsMySQL(cfg, ciliumLogs)
	} else if cfg.DBDriver == "sqlite3" {
		err = UpdateOrInsertCiliumLogsSQLite(cfg, ciliumLogs)
	}
	return err
}

func GetCiliumLogs(cfg types.ConfigDB, ciliumFilter types.CiliumLog) ([]types.CiliumLog, []uint32, error) {
	ciliumLogs := []types.CiliumLog{}
	ciliumTotalCount := []uint32{}
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		ciliumLogs, ciliumTotalCount, err = GetCiliumLogsMySQL(cfg, ciliumFilter)
	} else if cfg.DBDriver == "sqlite3" {
		ciliumLogs, ciliumTotalCount, err = GetCiliumLogsSQLite(cfg, ciliumFilter)
	}
	return ciliumLogs, ciliumTotalCount, err
}

func GetPodNames(cfg types.ConfigDB, filter types.ObsPodDetail) ([]string, error) {
	res := []string{}
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		res, err = GetPodNamesMySQL(cfg, filter)
	} else if cfg.DBDriver == "sqlite3" {
		res, err = GetPodNamesSQLite(cfg, filter)
	}
	return res, err
}

// =============== //
// == Policy DB == //
// =============== //
func GetPolicyYamls(cfg types.ConfigDB, policyType string) ([]types.PolicyYaml, error) {
	var err error
	var results []types.PolicyYaml

	if cfg.DBDriver == "mysql" {
		results, err = GetPolicyYamlsMySQL(cfg, policyType)
		if err != nil {
			return nil, err
		}
	} else if cfg.DBDriver == "sqlite3" {
		results, err = GetPolicyYamlsSQLite(cfg, policyType)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func UpdateOrInsertPolicyYamls(cfg types.ConfigDB, policies []types.PolicyYaml) error {
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		err = UpdateOrInsertPolicyYamlsMySQL(cfg, policies)
	} else if cfg.DBDriver == "sqlite3" {
		err = UpdateOrInsertPolicyYamlsSQLite(cfg, policies)
	}
	return err
}

// ============= //
// == Summary == //
// ============= //
func UpsertSystemSummary(cfg types.ConfigDB, summaryMap map[types.SystemSummary]types.SysSummaryTimeCount) error {
	var err = errors.New("unknown db driver")
	if cfg.DBDriver == "mysql" {
		err = UpsertSystemSummaryMySQL(cfg, summaryMap)
	} else if cfg.DBDriver == "sqlite3" {
		err = UpsertSystemSummarySQLite(cfg, summaryMap)
	}
	return err
}

func upsertSysSummarySQL(db *sql.DB, summary types.SystemSummary, timeCount types.SysSummaryTimeCount) error {
	queryString := `cluster_name = ? and cluster_id = ? and workspace_id = ? and namespace_name = ? and namespace_id = ? and container_name = ? and container_image = ? 
					and container_id = ? and podname = ? and operation = ? and labels = ? and deployment_name = ? and source = ? and destination = ? 
					and destination_namespace = ? and destination_labels = ? and type = ? and ip = ? and port = ? and protocol = ? and action = ?`

	query := "UPDATE " + TableSystemSummarySQLite + " SET count=count+?, updated_time=? WHERE " + queryString + " "

	updateStmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer updateStmt.Close()

	result, err := updateStmt.Exec(
		timeCount.Count,
		timeCount.UpdatedTime,
		summary.ClusterName,
		summary.ClusterId,
		summary.WorkspaceId,
		summary.NamespaceName,
		summary.NamespaceId,
		summary.ContainerName,
		summary.ContainerImage,
		summary.ContainerID,
		summary.PodName,
		summary.Operation,
		summary.Labels,
		summary.Deployment,
		summary.Source,
		summary.Destination,
		summary.DestNamespace,
		summary.DestLabels,
		summary.NwType,
		summary.IP,
		summary.Port,
		summary.Protocol,
		summary.Action,
	)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err == nil && rowsAffected == 0 {

		insertQueryString := `(cluster_name,cluster_id,workspace_id,namespace_name,namespace_id,container_name,container_image,container_id,podname,operation,labels,deployment_name,
				source,destination,destination_namespace,destination_labels,type,ip,port,protocol,action,count,updated_time) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

		insertQuery := "INSERT INTO " + TableSystemSummarySQLite + insertQueryString

		insertStmt, err := db.Prepare(insertQuery)
		if err != nil {
			return err
		}
		defer insertStmt.Close()

		_, err = insertStmt.Exec(
			summary.ClusterName,
			summary.ClusterId,
			summary.WorkspaceId,
			summary.NamespaceName,
			summary.NamespaceId,
			summary.ContainerName,
			summary.ContainerImage,
			summary.ContainerID,
			summary.PodName,
			summary.Operation,
			summary.Labels,
			summary.Deployment,
			summary.Source,
			summary.Destination,
			summary.DestNamespace,
			summary.DestLabels,
			summary.NwType,
			summary.IP,
			summary.Port,
			summary.Protocol,
			summary.Action,
			timeCount.Count,
			timeCount.UpdatedTime)
		if err != nil {
			log.Error().Msg(err.Error())
			return err
		}
	}

	return nil
}

func GetSystemSummary(cfg types.ConfigDB, filterOptions types.SystemSummary) ([]types.SystemSummary, error) {
	var err = errors.New("unknown db driver")
	res := []types.SystemSummary{}

	if cfg.DBDriver == "mysql" {
		res, err = GetSystemSummaryMySQL(cfg, filterOptions)
	} else if cfg.DBDriver == "sqlite3" {
		res, err = GetSystemSummarySQLite(cfg, filterOptions)
	}

	return res, err
}

func getSysSummarySQL(db *sql.DB, dbName string, filterOptions types.SystemSummary) ([]types.SystemSummary, error) {
	var results *sql.Rows
	var err error

	resSummary := []types.SystemSummary{}

	query := `SELECT cluster_name,cluster_id,workspace_id,namespace_name,namespace_id,container_name,
	container_image,container_id,podname,operation,labels,deployment_name,source,destination,destination_namespace,
	destination_labels,type,ip,port,protocol,action,count,updated_time FROM ` + dbName

	var whereClause string
	var args []interface{}

	if filterOptions.ClusterName != "" {
		concatWhereClause(&whereClause, "cluster_name")
		args = append(args, filterOptions.ClusterName)
	}
	if filterOptions.ClusterId != "" {
		concatWhereClause(&whereClause, "cluster_id")
		args = append(args, filterOptions.ClusterId)
	}
	if filterOptions.WorkspaceId != 0 {
		concatWhereClause(&whereClause, "workpsace_id")
		args = append(args, filterOptions.WorkspaceId)
	}
	if filterOptions.NamespaceName != "" {
		concatWhereClause(&whereClause, "namespace_name")
		args = append(args, filterOptions.NamespaceName)
	}
	if filterOptions.NamespaceId != 0 {
		concatWhereClause(&whereClause, "namespace_id")
		args = append(args, filterOptions.NamespaceId)
	}
	if filterOptions.ContainerName != "" {
		concatWhereClause(&whereClause, "container_name")
		args = append(args, filterOptions.ContainerName)
	}
	if filterOptions.ContainerImage != "" {
		concatWhereClause(&whereClause, "container_image")
		args = append(args, filterOptions.ContainerImage)
	}
	if filterOptions.ContainerID != "" {
		concatWhereClause(&whereClause, "container_id")
		args = append(args, filterOptions.ContainerID)
	}
	if filterOptions.PodName != "" {
		concatWhereClause(&whereClause, "podname")
		args = append(args, filterOptions.PodName)
	}
	if filterOptions.Operation != "" {
		concatWhereClause(&whereClause, "operation")
		args = append(args, filterOptions.Operation)
	}
	if filterOptions.Labels != "" {
		concatWhereClause(&whereClause, "labels")
		args = append(args, filterOptions.Labels)
	}
	if filterOptions.Deployment != "" {
		concatWhereClause(&whereClause, "deployment_name")
		args = append(args, filterOptions.Deployment)
	}
	if filterOptions.Source != "" {
		concatWhereClause(&whereClause, "source")
		args = append(args, filterOptions.Source)
	}
	if filterOptions.Destination != "" {
		concatWhereClause(&whereClause, "destination")
		args = append(args, filterOptions.Destination)
	}

	results, err = db.Query(query+whereClause, args...)

	if err != nil {
		log.Error().Msg(err.Error())
		return nil, err
	}
	defer results.Close()

	for results.Next() {
		localSum := types.SystemSummary{}
		if err := results.Scan(
			&localSum.ClusterName,
			&localSum.ClusterId,
			&localSum.WorkspaceId,
			&localSum.NamespaceName,
			&localSum.NamespaceId,
			&localSum.ContainerName,
			&localSum.ContainerImage,
			&localSum.ContainerID,
			&localSum.PodName,
			&localSum.Operation,
			&localSum.Labels,
			&localSum.Deployment,
			&localSum.Source,
			&localSum.Destination,
			&localSum.DestNamespace,
			&localSum.DestLabels,
			&localSum.NwType,
			&localSum.IP,
			&localSum.Port,
			&localSum.Protocol,
			&localSum.Action,
			&localSum.Count,
			&localSum.UpdatedTime,
		); err != nil {
			return nil, err
		}
		resSummary = append(resSummary, localSum)
	}

	return resSummary, err
}

// ==================================== //
// == Purge Old DB Entries Cron Job ==  //
// ==================================== //
var (
	CfgDB                                      types.ConfigDB
	SystemLogsMutex, NetworkLogsMutex, DBMutex *sync.Mutex
	PurgeDBCronJob                             *cron.Cron
	PurgeDBMap                                 types.ConfigPurgeOldDBEntries
)

func initMutex() {
	DBMutex = &sync.Mutex{}
	if cfg.GetCfgPurgeOldDBEntriesEnable() {
		SystemLogsMutex = &sync.Mutex{}
	}
	if cfg.GetCfgPurgeOldDBEntriesEnable() {
		NetworkLogsMutex = &sync.Mutex{}
	}
}

func InitPurgeOldDBEntries() {
	log = logger.GetInstance()
	CfgDB = cfg.GetCfgDB()

	if cfg.GetCfgPurgeOldDBEntriesEnable() {
		// Init mutex
		initMutex()

		PurgeDBCronJob = cron.New()
		err := PurgeDBCronJob.AddFunc(cfg.GetCfgPurgeOldDBEntriesCronJobTime(), PurgeOldDBEntriesCronJob) // time interval
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}
		PurgeDBCronJob.Start()
		log.Info().Msg("Purging Old DB Entries cron job started")
	}
}

func PurgeOldDBEntriesCronJob() {
	if cfg.GetCfgPurgeOldDBEntriesEnable() {
		go PurgeOldDBEntriesMySQL(types.ConfigDB{})
		go PurgeOldDBEntriesSQLite(types.ConfigDB{})
	}
}
