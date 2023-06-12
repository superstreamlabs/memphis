package server

import (
	"encoding/json"
	"memphis/db"
)

type CreateSchemaReq struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	CreatedByUsername string `json:"created_by_username"`
	SchemaContent     string `json:"schema_content"`
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
			err = updateSchemaVersion(existedSchema, csr)
			if err != nil {
				s.Errorf("createSchemaDirect: Schema " + csr.Name + ": " + err.Error())
			}
			return
		} else {
			s.Errorf("createSchemaDirect: Schema " + csr.Name + ": Bad Schema Type")
			return
		}
	}

	err = createNewSchemaDirect(csr, tenantName)
	if err != nil {
		s.Errorf("createSchemaDirect: Schema " + csr.Name + ": " + err.Error())
	}
	return
}

func updateSchemaVersion(existedSchema *Schema, newSchemaReq CreateSchemaReq) error {

}

func createNewSchemaDirect(newSchemaReq CreateSchemaReq, tenantName string) error {
	schemaVersionNumber := 1
	newSchema, rowUpdated, err = db.InsertNewSchema(newSchemaReq.Name, newSchemaReq.Type, newSchemaReq.CreatedByUsername, tenantName)

	exist, user, err := GetUserByUsername(newSchemaReq.CreatedByUsername, tenantName)
	if rowUpdated == 1 {
		_, _, err := db.InsertNewSchemaVersion(schemaVersionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, newSchema.ID, ///add missing fields)
	}
}
