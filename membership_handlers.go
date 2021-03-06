package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type membershipHandler struct {
	membershipService membershipService
}

func newMembershipHandler(ms membershipService) membershipHandler {
	return membershipHandler{
		membershipService: ms,
	}
}

func (mh *membershipHandler) refreshMembershipCache(writer http.ResponseWriter, req *http.Request) {
	err := mh.membershipService.refreshMembershipCache()
	if err != nil {
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writeJSONMessage(writer, "Memberships fetched", http.StatusOK)
	}
}

func (mh *membershipHandler) getMembershipsCount(writer http.ResponseWriter, req *http.Request) {
	err := mh.membershipService.refreshMembershipCache()
	if err != nil {
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
	} else {
		c := mh.membershipService.getMembershipCount()
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf(`%v`, c))
		buffer.WriteTo(writer)
	}
}

func (mh *membershipHandler) getMembershipUuids(writer http.ResponseWriter, req *http.Request) {
	uuids := mh.membershipService.getMembershipUuids()
	writeStreamResponse(uuids, writer)
}

func (mh *membershipHandler) getMembershipByUuid(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	m := mh.membershipService.getMembershipByUuid(uuid)
	writeJSONResponse(m, !reflect.DeepEqual(m, membership{}), writer)
}

func (mh *membershipHandler) AuthorsHealthCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to respond to request for curated author data from Bertha",
		Name:             "Check connectivity to Bertha Authors Spreadsheet",
		PanicGuide:       "https://dewey.in.ft.com/view/system/curated-authors-memberships-tf",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Bertha to be able to supply curated authors information",
		Checker:          mh.authorsChecker,
	}
}

func (mh *membershipHandler) authorsChecker() (string, error) {
	err := mh.membershipService.checkAuthorsConnectivity()
	if err == nil {
		return "Connectivity to Bertha Authors is ok", err
	}
	return "Error connecting to Bertha Authors", err
}

func (mh *membershipHandler) RolesHealthCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Unable to respond to request for curated author roles data from Bertha",
		Name:             "Check connectivity to Bertha Roles Spreadsheet",
		PanicGuide:       "https://dewey.in.ft.com/view/system/curated-authors-memberships-tf",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Bertha to be able to supply author roles",
		Checker:          mh.rolesChecker,
	}
}

func (mh *membershipHandler) rolesChecker() (string, error) {
	err := mh.membershipService.checkRolesConnectivity()
	if err == nil {
		return "Connectivity to Bertha Authors is ok", err
	}
	return "Error connecting to Bertha Authors", err
}

func (mh *membershipHandler) GTG() gtg.Status {
	rolesStatusCheck := func() gtg.Status {
		return gtgCheck(mh.rolesChecker)
	}

	authorsStatusCheck := func() gtg.Status {
		return gtgCheck(mh.authorsChecker)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{rolesStatusCheck, authorsStatusCheck})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

func writeJSONResponse(obj interface{}, found bool, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	if !found {
		writeJSONMessage(writer, "Membership not found", http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(writer)
	if err := enc.Encode(obj); err != nil {
		log.Errorf("Error on json encoding=%v\n", err)
		writeJSONMessage(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSONMessage(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, fmt.Sprintf("{\"message\": \"%s\"}", errorMsg))
}

func writeStreamResponse(ids []string, writer http.ResponseWriter) {
	for _, id := range ids {
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("{\"id\":\"%s\"}\n", id))
		buffer.WriteTo(writer)
	}
}
