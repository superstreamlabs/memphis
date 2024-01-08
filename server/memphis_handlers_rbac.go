package server

import (
	"fmt"
	"strings"

	"github.com/memphisdev/memphis/db"
	"github.com/memphisdev/memphis/models"
)

const (
	inboxSubject   = "_INBOX.>"
	wildCardSuffix = ".>"
)

// the function returns a bool for is allowd to create and a bool for if a reload is needed
func ValidateStationPermissions(rolesId []int, stationName, tenantName string) (bool, bool, error) {
	// if the user dosent have a role len rolesId is 0 then he allowd to create
	// TODO: add check if denied when we allow to deny
	if len(rolesId) == 0 {
		neededReload, err := checkTenantPermissionsUsage(tenantName)
		if err != nil {
			return false, false, err
		}
		return true, neededReload, nil
	} else {
		allowd, err := db.CheckUserStationPermissions(rolesId, stationName)
		if err != nil {
			return false, false, err
		}
		if !allowd {
			return false, false, nil
		}
		return true, true, nil
	}
}

func checkTenantPermissionsUsage(tenantName string) (bool, error) {
	//if the tenant use rbac than reload is needed
	reloadNeeded, err := db.CheckTenantPermissionsUsage(tenantName)
	if err != nil {
		return false, err
	}
	return reloadNeeded, nil
}

func GetUserAllowedStations(userRoles []int, tenantName string) ([]models.Station, []models.Station, error) {
	permissions, err := db.GetUserPermissions(userRoles, tenantName)
	if err != nil {
		return nil, nil, err
	}

	var permissionsAllowReadPattern []string
	var permissionsAllowWritePattern []string

	for _, permission := range permissions {
		if permission.RestrictionType == "allow" {
			if permission.Type == "read" {
				permissionsAllowReadPattern = append(permissionsAllowReadPattern, permission.Pattern)

			} else if permission.Type == "write" {
				permissionsAllowWritePattern = append(permissionsAllowWritePattern, permission.Pattern)
			}
		}
	}

	allowReadStations, err := db.GetStationsByPattern(permissionsAllowReadPattern, tenantName)
	if err != nil {
		return nil, nil, err
	}

	allowWriteStations, err := db.GetStationsByPattern(permissionsAllowWritePattern, tenantName)
	if err != nil {
		return nil, nil, err
	}

	return allowReadStations, allowWriteStations, nil
}

func GetPatternWithDots(pattern string) string {
	return strings.Replace(pattern, ".", "\\.\\\\", -1)
}

func GetAllInternalSbjectsForWriteRespones(station models.Station) []string {
	var subjects []string

	subjects = append(subjects, fmt.Sprintf(schemaUpdatesSubjectTemplate, replaceDelimiters(station.Name)))
	subjects = append(subjects, fmt.Sprintf(FUNCTIONS_UPDATE_SUBJ, replaceDelimiters(station.Name)))

	return subjects
}

func GetAllowedSubjectsFromRoleIds(roleIds []int, tenantName string) ([]string, []string, error) {
	allowReadStations, allowdWriteStations, err := GetUserAllowedStations(roleIds, tenantName)
	if err != nil {
		return nil, nil, err
	}

	var allowReadSubjects []string
	var allowWriteSubjects []string

	for _, station := range allowReadStations {
		for _, partition := range station.PartitionsList {
			partitionStream := fmt.Sprintf("%v$%v.>", replaceDelimiters(station.Name), partition)
			allowReadSubjects = append(allowReadSubjects, GetAllowReadSubscribeInternalSubjects(partitionStream)...)
			allowWriteSubjects = append(allowWriteSubjects, GetAllowReadPublishInternalSbjects(partitionStream)...)
		}
		allowReadSubjects = append(allowReadSubjects, GetAllMemphisStationInternalSubjects(station.Name)...)
		allowWriteSubjects = append(allowWriteSubjects, GetAllMemphisStationInternalSubjects(station.Name)...)
	}

	for _, station := range allowdWriteStations {
		for _, partition := range station.PartitionsList {
			partitionStream := fmt.Sprintf("%v$%v.>", replaceDelimiters(station.Name), partition)
			allowWriteSubjects = append(allowWriteSubjects, GetAllowWritePublishInternalSubjects(partitionStream)...)
		}
		allowWriteSubjects = append(allowWriteSubjects, GetAllMemphisStationInternalSubjects(station.Name)...)
		allowReadSubjects = append(allowReadSubjects, GetAllMemphisStationInternalSubjects(station.Name)...)
	}

	allowReadSubjects = append(allowReadSubjects, GetAllMemphisAndNatsInternalSubjects()...)
	allowWriteSubjects = append(allowWriteSubjects, GetAllMemphisAndNatsInternalSubjects()...)

	return allowReadSubjects, allowWriteSubjects, nil
}

func GetAllMemphisAndNatsInternalSubjects() []string {
	var subjects []string

	// Memphis subjects
	subjects = append(subjects, SCHEMAVERSE_DLS_SUBJ)
	subjects = append(subjects, sdkClientsUpdatesSubject)
	subjects = append(subjects, PM_RESEND_ACK_SUBJ)
	subjects = append(subjects, memphisSchemaDetachments)
	subjects = append(subjects, memphisConsumerCreations)
	subjects = append(subjects, memphisConsumerDestructions)
	subjects = append(subjects, memphisNotifications)
	subjects = append(subjects, memphisProducerCreations)
	subjects = append(subjects, memphisProducerDestructions)
	subjects = append(subjects, memphisSchemaCreations)
	subjects = append(subjects, memphisStationCreations)
	subjects = append(subjects, memphisStationDestructions)

	// Nats subjects
	subjects = append(subjects, inboxSubject)
	subjects = append(subjects, JSApiStreams)
	subjects = append(subjects, JSApiAccountInfo)

	return subjects
}

func GetAllMemphisStationInternalSubjects(stationName string) []string {
	var subjects []string

	// Memphis subjects
	subjects = append(subjects, fmt.Sprintf(FUNCTIONS_UPDATE_SUBJ, replaceDelimiters(stationName)))
	subjects = append(subjects, fmt.Sprintf(connectConfigUpdatesSubjectTemplate, replaceDelimiters(stationName)))
	subjects = append(subjects, fmt.Sprintf(schemaUpdatesSubjectTemplate, replaceDelimiters(stationName)))
	subjects = append(subjects, fmt.Sprintf(memphisWS_TemplSubj_Publish, replaceDelimiters(stationName)))
	subjects = append(subjects, fmt.Sprintf(dlsResendMessagesStreamNew, replaceDelimiters(stationName), ">"))

	return subjects
}

func GetAllowReadSubscribeInternalSubjects(partitionStream string) []string {
	var subjects []string
	subjects = append(subjects, partitionStream)
	return subjects
}

func GetAllowReadPublishInternalSbjects(partitionStream string) []string {
	var subjects []string

	// Nats subjects
	subjects = append(subjects, fmt.Sprintf("%v%v", jsAckPre, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiConsumerListT, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiConsumerCreateT, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiMsgGetT, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiRequestNextTMemphis, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiConsumerDeleteTMemphis, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiDurableCreateTMemphis, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiConsumerInfoTMemphis, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiStreamInfoT, strings.TrimSuffix(partitionStream, wildCardSuffix)))

	return subjects
}

func GetAllowWritePublishInternalSubjects(partitionStream string) []string {
	var subjects []string

	subjects = append(subjects, partitionStream)
	subjects = append(subjects, fmt.Sprintf(JSApiStreamCreateT, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiStreamDeleteT, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiStreamUpdateT, partitionStream))
	subjects = append(subjects, fmt.Sprintf(JSApiStreamInfoT, strings.TrimSuffix(partitionStream, wildCardSuffix)))

	return subjects
}
