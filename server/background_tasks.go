package server

import (
	"encoding/json"
	"errors"
	"memphis-broker/models"
	"memphis-broker/notifications"
	"strings"
	"time"
)

const CONN_STATUS_SUBJ = "$memphis_connection_status"
const INTEGRATIONS_UPDATES_SUBJ = "$memphis_integration_updates"
const SCHEMA_VALIDATION_FAIL_SUBJ = "$memphis_schema_validation_fail_updates"

func (s *Server) ListenForZombieConnCheckRequests() error {
	_, err := s.subscribeOnGlobalAcc(CONN_STATUS_SUBJ, CONN_STATUS_SUBJ+"_sid", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			message := strings.TrimSuffix(string(msg), "\r\n")
			reported := checkAndReportConnFound(s, message, reply)

			if !reported {
				maxIterations := 14
				for range time.Tick(time.Second * 2) {
					reported = checkAndReportConnFound(s, message, reply)
					if reported {
						return
					}
					maxIterations--
					if maxIterations == 0 {
						return
					}
				}
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func checkAndReportConnFound(s *Server, message, reply string) bool {
	connInfo := &ConnzOptions{}
	conns, _ := s.Connz(connInfo)
	for _, conn := range conns.Conns {
		connId := strings.Split(conn.Name, "::")[0]
		if connId == message {
			s.sendInternalAccountMsgWithReply(s.GlobalAccount(), reply, _EMPTY_, nil, []byte("connExists"), true)
			return true
		}
	}
	return false
}

func (s *Server) ListenForIntegrationsUpdateEvents() error {
	_, err := s.subscribeOnGlobalAcc(INTEGRATIONS_UPDATES_SUBJ, INTEGRATIONS_UPDATES_SUBJ+"_sid"+s.Name(), func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var integrationUpdate models.Integration
			err := json.Unmarshal(msg, &integrationUpdate)
			if err != nil {
				s.Errorf(err.Error())
			}
			switch strings.ToLower(integrationUpdate.Name) {
			case "slack":
				notifications.CacheSlackDetails(integrationUpdate.Keys, integrationUpdate.Properties, integrationUpdate.UIUrl)
			default:
				return
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) ListenForSchemaValidationFailEvents() error {
	err := s.queueSubscribe(SCHEMA_VALIDATION_FAIL_SUBJ, SCHEMA_VALIDATION_FAIL_SUBJ+"_group", func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var schemaFailMsg models.SchemaFailMsg
			err := json.Unmarshal(msg, &schemaFailMsg)
			if err != nil {
				return
			}
			slackIntegration, ok := notifications.NotificationIntegrationsMap["slack"].(models.SlackIntegration)
			if !ok {
				return
			}
			if slackIntegration.Properties["schema_validation_fail_alert"] {
				notifications.SendMessageToSlackChannel(schemaFailMsg.Title, schemaFailMsg.Msg)
			}
		}(copyBytes(msg))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) StartBackgroundTasks() error {
	err := s.ListenForZombieConnCheckRequests()
	if err != nil {
		return errors.New("Failed subscribing for zombie conns check requests: " + err.Error())
	}

	err = s.ListenForIntegrationsUpdateEvents()
	if err != nil {
		return errors.New("Failed subscribing for integrations updates: " + err.Error())
	}

	err = s.ListenForSchemaValidationFailEvents()
	if err != nil {
		return errors.New("Failed subscribing for schema validation updates: " + err.Error())
	}
	return nil
}
