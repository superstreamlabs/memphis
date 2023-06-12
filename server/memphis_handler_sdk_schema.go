package server

import (
	"encoding/json"
	"memphis/db"
	"strconv"
)

type CreateSchemaReq struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	CreatedByUsername string `json:"created_by_username"`
	SchemaContent     string `json:"schema_content"`
	MessageStructName string `json:"message_struct_name"`
}

func (s *Server) createSchemaDirect(c *client, reply string, msg []byte) {
	var csr CreateSchemaReq
	var tenantName string
	tenantName, message, err := s.getTenantNameAndMessage(msg) // will it work???
	if err != nil {
		s.Errorf("createSchemaDirect: " + err.Error())
		return
	}
	if err := json.Unmarshal([]byte(message), &csr); err != nil {
		s.Errorf("createSchemaDirect: failed creating Schema: %v", err.Error())
		respondWithErr(globalAccountName, s, reply, err)
		return
	}

	exist, existedSchema, err := db.GetSchemaByName(csr.Name, tenantName)
	if err != nil {
		s.Errorf("createSchemaDirect: Schema " + csr.Name + ": " + err.Error())
	}

	if exist {
		if existedSchema.Type == csr.Type {
			s.updateSchemaVersion(existedSchema.ID, tenantName, csr)
			return
		} else {
			s.Errorf("createSchemaDirect: Schema " + csr.Name + ": Bad Schema Type")
			return
		}
	}

	s.createNewSchemaDirect(csr, tenantName)

}

func (s *Server) updateSchemaVersion(schemaID int, tenantName string, newSchemaReq CreateSchemaReq) {
	_, user, err := db.GetUserByUsername(newSchemaReq.CreatedByUsername, tenantName)

	countVersions, err := db.GetShcemaVersionsCount(schemaID, user.TenantName)
	if err != nil {
		s.Errorf("updateSchemaVersion: Schema " + newSchemaReq.Name + ": " + err.Error())
		return
	}

	versionNumber := countVersions + 1

	descriptor := ""
	if newSchemaReq.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(newSchemaReq.Name, 1, newSchemaReq.SchemaContent, newSchemaReq.Type)
		if err != nil {
			s.Warnf("CreateNewSchemaDirectn: Schema " + newSchemaReq.Name + ": " + err.Error())
			return
		}
	}

	newSchemaVersion, rowsUpdated, err := db.InsertNewSchemaVersion(versionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, schemaID, newSchemaReq.MessageStructName, descriptor, false, tenantName)
	if err != nil {
		s.Warnf("updateSchemaVersion: " + err.Error())
		return
	}
	if rowsUpdated == 1 {
		message := "Schema Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " has been created by " + user.Username
		s.Noticef(message)
	} else {
		s.Warnf("CreateNewVersion: Schema " + newSchemaReq.Name + ": Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " already exists")
		return
	}

}

func (s *Server) createNewSchemaDirect(newSchemaReq CreateSchemaReq, tenantName string) {
	schemaVersionNumber := 1

	_, user, err := db.GetUserByUsername(newSchemaReq.CreatedByUsername, tenantName)

	descriptor := ""
	if newSchemaReq.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(newSchemaReq.Name, 1, newSchemaReq.SchemaContent, newSchemaReq.Type)
		if err != nil {
			s.Warnf("CreateNewSchemaDirectn: Schema " + newSchemaReq.Name + ": " + err.Error())
			return
		}
	}

	newSchema, rowUpdated, err := db.InsertNewSchema(newSchemaReq.Name, newSchemaReq.Type, newSchemaReq.CreatedByUsername, tenantName)

	if rowUpdated == 1 {
		_, _, err := db.InsertNewSchemaVersion(schemaVersionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, newSchema.ID, newSchemaReq.MessageStructName, descriptor, true, tenantName)
		if err != nil {
			s.Errorf("createNewSchemaDirect: " + err.Error())
			return
		}
	}
}
