package server

import (
	"encoding/json"
	"memphis-broker/models"
	"memphis-broker/notifications"
	"strings"
	"time"
)

const INTEGRATIONS_UPDATES_SUBJ = "$memphis_integration_updates"

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

func (s *Server) ListenForIntegrationsUpdates() error {
	_, err := s.subscribeOnGlobalAcc(INTEGRATIONS_UPDATES_SUBJ, INTEGRATIONS_UPDATES_SUBJ+"_sid"+s.Name(), func(_ *client, subject, reply string, msg []byte) {
		go func(msg []byte) {
			var integrationUpdate models.Integration
			err := json.Unmarshal(msg, &integrationUpdate)
			if err != nil {
				s.Errorf(err.Error())
			}
			switch integrationUpdate.Name {
			case "slack":
				notifications.UpdateSlackDetails(integrationUpdate.Keys, integrationUpdate.Properties, integrationUpdate.UIUrl)
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
