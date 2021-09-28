package libs

import (
	"strings"
	"time"

	"github.com/accuknox/knoxAutoPolicy/src/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ================= //
// == Network Log == //
// ================= //

// LastFlowID network flow between [ startTime <= time < endTime ]
var LastFlowID int64 = 0
var startTime int64 = 0
var endTime int64 = 0

func updateTimeInterval(lastDoc map[string]interface{}) {
	if val, ok := lastDoc["timestamp"].(primitive.DateTime); ok {
		ts := val
		startTime = ts.Time().Unix() + 1
	} else if val, ok := lastDoc["timestamp"].(uint32); ok {
		startTime = int64(val) + 1
	}
}

func GetNetworkLogsFromDB(cfg types.ConfigDB, timeSelection string, trigger, limit int) []map[string]interface{} {
	results := []map[string]interface{}{}

	endTime = time.Now().Unix()

	if cfg.DBDriver == "mysql" {
		if timeSelection == "" {
			docs, err := GetNetworkLogByIDTimeFromMySQL(cfg, LastFlowID, endTime, limit)
			if err != nil {
				log.Error().Msg(err.Error())
				return results
			}
			results = docs

			if len(results) != 0 && len(results) < trigger {
				log.Info().Msgf("The number of network logs [%d] is less than trigger [%d]", len(results), trigger)
				return results
			}
		} else {
			// given time selection from ~ to
			times := strings.Split(timeSelection, "|")
			from := ConvertStrToUnixTime(times[0])
			to := ConvertStrToUnixTime(times[1])

			docs, err := GetNetworkLogByTimeFromMySQL(cfg, from, to)
			if err != nil {
				log.Error().Msg(err.Error())
				return results
			}
			results = docs
		}
	} else {
		return results
	}

	if len(results) == 0 {
		log.Info().Msgf("Network logs not exist: from %s ~ to %s",
			time.Unix(startTime, 0).Format(TimeFormSimple),
			time.Unix(endTime, 0).Format(TimeFormSimple))

		return results
	}

	lastDoc := results[len(results)-1]

	// id update for mysql
	if cfg.DBDriver == "mysql" {
		LastFlowID = int64(lastDoc["id"].(uint32))
	}

	log.Info().Msgf("The total number of network logs: [%d] from %s ~ to %s", len(results),
		time.Unix(startTime, 0).Format(TimeFormSimple),
		time.Unix(endTime, 0).Format(TimeFormSimple))

	startTime = endTime + 1

	return results
}

func InsertNetworkLogToDB(cfg types.ConfigDB, nfe []types.NetworkLogEvent) error {
	if cfg.DBDriver == "mysql" {
		if err := InsertNetworkLogToMySQL(cfg, nfe); err != nil {
			return err
		}
	}

	return nil
}

// ==================== //
// == Network Policy == //
// ==================== //

func GetNetworkPolicies(cfg types.ConfigDB, cluster, namespace, status string) []types.KnoxNetworkPolicy {
	results := []types.KnoxNetworkPolicy{}

	if cfg.DBDriver == "mysql" {
		docs, err := GetNetworkPoliciesFromMySQL(cfg, cluster, namespace, status)
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
		docs, err := GetNetworkPoliciesFromMySQL(cfg, cluster, namespace, status)
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
			val = policy.Spec.Selector.MatchLabels[k]
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
	}
}

func InsertNetworkPolicies(cfg types.ConfigDB, policies []types.KnoxNetworkPolicy) {
	if cfg.DBDriver == "mysql" {
		if err := InsertNetworkPoliciesToMySQL(cfg, policies); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

// ================ //
// == System Log == //
// ================ //

// LastSyslogID system log between [ startTime <= time < endTime ]
var LastSyslogID int64 = 0
var syslogStartTime int64 = 0
var syslogEndTime int64 = 0

func GetSystemLogsFromDB(cfg types.ConfigDB, timeSelection string, trigger, limit int) []map[string]interface{} {
	results := []map[string]interface{}{}

	syslogEndTime = time.Now().Unix()

	if cfg.DBDriver == "mysql" {
		if timeSelection == "" {
			docs, err := GetSystemLogByIDTimeFromMySQL(cfg, LastSyslogID, syslogEndTime, limit)
			if err != nil {
				log.Error().Msg(err.Error())
				return results
			}
			results = docs

			if len(results) != 0 && len(results) < trigger {
				log.Info().Msgf("The number of system logs [%d] is less than trigger [%d]", len(results), trigger)
				return results
			}
		} else {
			// given time selection from ~ to
			times := strings.Split(timeSelection, "|")
			from := ConvertStrToUnixTime(times[0])
			to := ConvertStrToUnixTime(times[1])

			docs, err := GetSystemLogByTimeFromMySQL(cfg, from, to)
			if err != nil {
				log.Error().Msg(err.Error())
				return results
			}
			results = docs
		}
	} else {
		return results
	}

	if len(results) == 0 {
		log.Info().Msgf("System logs not exist: from %s ~ to %s",
			time.Unix(syslogStartTime, 0).Format(TimeFormSimple),
			time.Unix(syslogEndTime, 0).Format(TimeFormSimple))

		return results
	}

	lastDoc := results[len(results)-1]

	// id update for mysql
	if cfg.DBDriver == "mysql" {
		LastSyslogID = int64(lastDoc["id"].(uint32))
	}

	log.Info().Msgf("The total number of system logs: [%d] from %s ~ to %s", len(results),
		time.Unix(syslogStartTime, 0).Format(TimeFormSimple),
		time.Unix(syslogEndTime, 0).Format(TimeFormSimple))

	syslogStartTime = syslogEndTime + 1

	return results
}

func InsertSystemLogToDB(cfg types.ConfigDB, sle []types.SystemLogEvent) error {
	if cfg.DBDriver == "mysql" {
		if err := InsertSystemLogToMySQL(cfg, sle); err != nil {
			return err
		}
	}

	return nil
}

// ================== //
// == System Alert == //
// ================== //

// LastSysAlertID system_alert between [ startTime <= time < endTime ]
var LastSysAlertID int64 = 0
var sysAlertStartTime int64 = 0
var sysAlertEndTime int64 = 0

func GetSystemAlertsFromDB(cfg types.ConfigDB, timeSelection string, trigger, limit int) []map[string]interface{} {
	results := []map[string]interface{}{}

	sysAlertEndTime = time.Now().Unix()

	if cfg.DBDriver == "mysql" {
		if timeSelection == "" {
			docs, err := GetSystemAlertByIDTimeFromMySQL(cfg, LastSysAlertID, sysAlertEndTime, limit)
			if err != nil {
				log.Error().Msg(err.Error())
				return results
			}
			results = docs

			// TOOD: checking alert
			// if len(results) != 0 && len(results) < trigger {
			// 	log.Info().Msgf("The number of system alerts [%d] is less than trigger [%d]", len(results), trigger)
			// 	return results
			// }
		} else {
			// given time selection from ~ to
			times := strings.Split(timeSelection, "|")
			from := ConvertStrToUnixTime(times[0])
			to := ConvertStrToUnixTime(times[1])

			docs, err := GetSystemAlertByTimeFromMySQL(cfg, from, to)
			if err != nil {
				log.Error().Msg(err.Error())
				return results
			}
			results = docs
		}
	} else {
		return results
	}

	if len(results) == 0 {
		log.Info().Msgf("System alerts not exist: from %s ~ to %s",
			time.Unix(sysAlertStartTime, 0).Format(TimeFormSimple),
			time.Unix(sysAlertEndTime, 0).Format(TimeFormSimple))

		return results
	}

	lastDoc := results[len(results)-1]

	// id update for mysql
	if cfg.DBDriver == "mysql" {
		LastSyslogID = int64(lastDoc["id"].(uint32))
	}

	log.Info().Msgf("The total number of system alerts: [%d] from %s ~ to %s", len(results),
		time.Unix(sysAlertStartTime, 0).Format(TimeFormSimple),
		time.Unix(sysAlertEndTime, 0).Format(TimeFormSimple))

	sysAlertStartTime = sysAlertEndTime + 1

	return results
}

func InsertSystemAlertToDB(cfg types.ConfigDB, sae []types.SystemAlertEvent) error {
	if cfg.DBDriver == "mysql" {
		if err := InsertSystemAlertToMySQL(cfg, sae); err != nil {
			return err
		}
	}

	return nil
}

// =================== //
// == System Policy == //
// =================== //

func UpdateOutdatedSystemPolicy(cfg types.ConfigDB, outdatedPolicy string, latestPolicy string) {
	if cfg.DBDriver == "mysql" {
		if err := UpdateOutdatedNetworkPolicyFromMySQL(cfg, outdatedPolicy, latestPolicy); err != nil {
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
	}

	return results
}

func InsertSystemPolicies(cfg types.ConfigDB, policies []types.KnoxSystemPolicy) {
	if cfg.DBDriver == "mysql" {
		if err := InsertSystemPoliciesToMySQL(cfg, policies); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

// =========== //
// == Table == //
// =========== //

func ClearDBTables(cfg types.ConfigDB) {
	if cfg.DBDriver == "mysql" {
		if err := ClearDBTablesMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func CreateTablesIfNotExist(cfg types.ConfigDB) {
	if cfg.DBDriver == "mysql" {
		if err := CreateTableNetworkLogMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableNetworkPolicyMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemLogMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemAlertMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
		if err := CreateTableSystemPolicyMySQL(cfg); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}
