package server

import (
	"encoding/json"
	"errors"
	"fmt"
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

type DeleteSchemaReq struct {
	Name string `json:"name"`
}

type SchemaResponse struct {
	Err string `json:"error"`
}

func (s *Server) deleteSchemaDirect(c *client, reply string, msg []byte) {
	var tenantName string
	var dsr DeleteSchemaReq
	var resp SchemaResponse
	tenantName, message, err := s.getTenantNameAndMessage(msg)
	if err != nil {
		s.Errorf("deleteSchemaDirect - failed deleting Schema: " + err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}

	if err := json.Unmarshal([]byte(message), &dsr); err != nil {
		s.Errorf("deleteSchemaDirect - failed deleting Schema: %v", err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}

	exist, schema, err := db.GetSchemaByName(dsr.Name, tenantName)
	if err != nil {
		s.Errorf("deleteSchemaDirect - failed deleting Schema:" + err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	} else if !exist {
		schemaDoesntExist := fmt.Sprintf(" %v  Doesn't Exist", dsr.Name)
		s.Errorf(schemaDoesntExist)
		respondWithRespErr(tenantName, s, reply, errors.New(schemaDoesntExist), &resp)
		return
	}

	err = db.FindAndDeleteSchema([]int{schema.ID})
	if err != nil {
		s.Errorf("deleteSchemaDirect - failed deleting Schema:" + err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}

	respondWithRespErr(tenantName, s, reply, err, &resp)

}

func (s *Server) createSchemaDirect(c *client, reply string, msg []byte) {
	var csr CreateSchemaReq
	var resp SchemaResponse
	var tenantName string
	tenantName, message, err := s.getTenantNameAndMessage(msg) // will it work???
	if err != nil {
		s.Errorf("createSchemaDirect - failed creating Schema:" + err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}
	if err := json.Unmarshal([]byte(message), &csr); err != nil {
		s.Errorf("createSchemaDirect - failed creating Schema: %v", err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}

	exist, existedSchema, err := db.GetSchemaByName(csr.Name, tenantName)
	if err != nil {
		s.Errorf("createSchemaDirect - failed creating Schema: " + err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}

	if exist {
		if existedSchema.Type == csr.Type {
			err = s.updateSchemaVersion(existedSchema.ID, tenantName, csr)
			if err != nil {
				s.Errorf("createSchemaDirect - failed creating Schema: " + csr.Name + err.Error())
				respondWithRespErr(tenantName, s, reply, err, &resp)
				return
			}
			respondWithRespErr(tenantName, s, reply, err, &resp)
			return
		} else {
			s.Errorf("createSchemaDirect: Schema " + csr.Name + ": Bad Schema Type")
			badTypeError := fmt.Sprintf("%v already exist with type - %v", csr.Name, existedSchema.Type)
			respondWithRespErr(tenantName, s, reply, errors.New(badTypeError), &resp)
			return
		}
	}

	err = s.createNewSchemaDirect(csr, tenantName)
	if err != nil {
		s.Errorf("createSchemaDirect - failed creating Schema:" + csr.Name + err.Error())
		respondWithRespErr(tenantName, s, reply, err, &resp)
		return
	}

	respondWithRespErr(tenantName, s, reply, err, &resp)

}

func (s *Server) updateSchemaVersion(schemaID int, tenantName string, newSchemaReq CreateSchemaReq) error {
	_, user, err := db.GetUserByUsername(newSchemaReq.CreatedByUsername, tenantName)
	if err != nil {
		s.Errorf("updateSchemaVersion: Schema " + newSchemaReq.Name + ": " + err.Error())
		return err
	}

	countVersions, err := db.GetShcemaVersionsCount(schemaID, user.TenantName)
	if err != nil {
		s.Errorf("updateSchemaVersion: Schema " + newSchemaReq.Name + ": " + err.Error())
		return err
	}

	_, currentSchema, err := db.GetSchemaVersionByNumberAndID(countVersions, schemaID)
	if err != nil {
		s.Errorf("updateSchemaVersion: Schema " + newSchemaReq.Name + ": " + err.Error())
		return err
	}

	if currentSchema.SchemaContent == newSchemaReq.SchemaContent {
		alreadyExistInDB := fmt.Sprintf("%v already exist in the db", newSchemaReq.Name)
		s.Errorf(alreadyExistInDB)
		return errors.New(alreadyExistInDB)
	}

	versionNumber := countVersions + 1

	descriptor := ""
	if newSchemaReq.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(newSchemaReq.Name, 1, newSchemaReq.SchemaContent, newSchemaReq.Type)
		if err != nil {
			s.Errorf("CreateNewSchemaDirectn: could not create proto descriptor for " + newSchemaReq.Name + ": " + err.Error())
			return err
		}
	}

	newSchemaVersion, rowsUpdated, err := db.InsertNewSchemaVersion(versionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, schemaID, newSchemaReq.MessageStructName, descriptor, false, tenantName)
	if err != nil {
		s.Errorf("updateSchemaVersion: " + err.Error())
		return err
	}
	if rowsUpdated == 1 {
		message := "Schema Version " + strconv.Itoa(newSchemaVersion.VersionNumber) + " has been created by " + user.Username
		s.Noticef(message)
		return nil
	} else {
		s.Errorf("updateSchemaVersion: schema update failed")
		return errors.New("updateSchemaVersion: schema update failed")
	}

}

func (s *Server) createNewSchemaDirect(newSchemaReq CreateSchemaReq, tenantName string) error {
	schemaVersionNumber := 1

	_, user, err := db.GetUserByUsername(newSchemaReq.CreatedByUsername, tenantName)

	descriptor := ""
	if newSchemaReq.Type == "protobuf" {
		descriptor, err = generateSchemaDescriptor(newSchemaReq.Name, 1, newSchemaReq.SchemaContent, newSchemaReq.Type)
		if err != nil {
			s.Errorf("CreateNewSchemaDirectn: Schema " + newSchemaReq.Name + ": " + err.Error())
			return err
		}
	}

	newSchema, rowUpdated, err := db.InsertNewSchema(newSchemaReq.Name, newSchemaReq.Type, newSchemaReq.CreatedByUsername, tenantName)
	if err != nil {
		s.Errorf("createNewSchemaDirect: " + err.Error())
		return err
	}

	if rowUpdated == 1 {
		_, _, err := db.InsertNewSchemaVersion(schemaVersionNumber, user.ID, user.Username, newSchemaReq.SchemaContent, newSchema.ID, newSchemaReq.MessageStructName, descriptor, true, tenantName)
		if err != nil {
			s.Errorf("createNewSchemaDirect: " + err.Error())
			return err
		}
	}

	return nil
}

func (csresp *SchemaResponse) SetError(err error) {
	if err != nil {
		csresp.Err = err.Error()
	} else {
		csresp.Err = ""
	}
}
