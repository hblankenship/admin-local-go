package shared

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var CP_base_url = "https://api.copper.com/developer_api/v1/"
var CP_projects_fragment = "projects/"
var CP_opp_fragment = "opportunities/"
var CP_pipeline_fragment = "pipelines/"
var CP_people_fragment = "people/"
var CP_related_fragment = ":entity/:entity_id/related"
var CP_custfields_fragment = "custom_field_definitions/"
var CP_search_fragment = "search"

// Custom Field Definition Ids
var CP_project_type = 399609
var CP_project_type_option_global_event = 899314
var CP_project_type_option_regional_event = 899315
var CP_project_type_option_chapter = 899316
var CP_project_type_option_global_partner = 900407
var CP_project_type_option_local_partner = 900408
var CP_project_type_option_project = 1378082
var CP_project_type_option_committee = 1378083
var CP_project_github_repo = 399740

// event specific
var CP_project_event_start_date = 392473
var CP_project_event_website = 395225
var CP_project_event_sponsorship_url = 395226
var CP_project_event_projected_revenue = 392478
var CP_project_event_sponsors = 392480
var CP_project_event_jira_ticket = 394290
var CP_project_event_approved_date = 392477

// chapter specific
var CP_project_chapter_status = 399736
var CP_project_chapter_status_option_active = 899462
var CP_project_chapter_status_option_inactive = 899463
var CP_project_chapter_status_option_suspended = 899464
var CP_project_chapter_region = 399739
var CP_project_chapter_region_option_africa = 899465
var CP_project_chapter_region_option_asia = 899466
var CP_project_chapter_region_option_centralamerica = 1607249
var CP_project_chapter_region_option_eastern_europe = 1607250
var CP_project_chapter_region_option_european_union = 899467
var CP_project_chapter_region_option_middle_east = 1607251
var CP_project_chapter_region_option_northamerica = 899468
var CP_project_chapter_region_option_oceania = 899469
var CP_project_chapter_region_option_southamerica = 899470
var CP_project_chapter_region_option_the_caribbean = 1607252
var CP_project_chapter_country = 399738
var CP_project_chapter_postal_code = 399737

// person specific
// inactive CP_person_group_url = 394184
// inactive CP_person_group_type = 394186
// inactive CP_person_group_type_option_chapter=672528
// inactive CP_person_group_type_option_project=672529
// inactive CP_person_group_type_option_committee=672530
// inactive CP_person_group_participant_type = 394187
// inactive CP_person_group_participant_type_option_participant = 672531
// inactive CP_person_group_participant_type_option_leader = 672532
// inactive CP_person_member_checkbox = 394880
// inactive CP_person_leader_checkbox = 394881
var CP_person_membership = 394882
var CP_person_membership_option_student = 674397
var CP_person_membership_option_lifetime = 674398
var CP_person_membership_option_oneyear = 674395
var CP_person_membership_option_twoyear = 674396
var CP_person_membership_option_complimentary = 1506889
var CP_person_membership_option_honorary = 1519960
var CP_person_membership_start = 394883
var CP_person_membership_end = 394884
var CP_person_github_username = 395220
var CP_person_signed_leaderagreement = 448262

// inactive CP_person_membership_number = 397651
var CP_person_external_id = 400845 //old Salesforce id
var CP_person_stripe_number = 440584

// /opportunity specific
var CP_opportunity_end_date = 400119
var CP_opportunity_autorenew_checkbox = 419575
var CP_opportunity_invoice_no = 407333 //can be the URL to the stripe payment for membership
var CP_opportunity_pipeline_id_membership = 721986
var CP_opportunity_stripe_transaction_id = 440903

type CopperPerson struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Prefix        interface{}   `json:"prefix"`
	FirstName     string        `json:"first_name"`
	MiddleName    interface{}   `json:"middle_name"`
	LastName      string        `json:"last_name"`
	Suffix        interface{}   `json:"suffix"`
	Address       CopperAddress `json:"address"`
	AssigneeID    interface{}   `json:"assignee_id"`
	CompanyID     interface{}   `json:"company_id"`
	CompanyName   interface{}   `json:"company_name"`
	ContactTypeID int           `json:"contact_type_id"`
	Details       interface{}   `json:"details"`
	Emails        []struct {
		Email    string `json:"email"`
		Category string `json:"category"`
	} `json:"emails"`
	PhoneNumbers []struct {
		Number   string `json:"number"`
		Category string `json:"category"`
	} `json:"phone_numbers"`
	Socials      []interface{} `json:"socials"`
	Tags         []string      `json:"tags"`
	Title        interface{}   `json:"title"`
	Websites     []interface{} `json:"websites"`
	CustomFields []struct {
		CustomFieldDefinitionID int         `json:"custom_field_definition_id"`
		Value                   interface{} `json:"value"`
	} `json:"custom_fields"`
	DateCreated      int `json:"date_created"`
	DateModified     int `json:"date_modified"`
	InteractionCount int `json:"interaction_count"`
}

type CopperAddress struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type CopperCustomFields []struct {
	CustomFieldDefinitionID int         `json:"custom_field_definition_id"`
	Value                   interface{} `json:"value"`
}

type Opportunities []struct {
	ID                 int                `json:"id"`
	Name               string             `json:"name"`
	AssigneeID         interface{}        `json:"assignee_id"`
	CloseDate          string             `json:"close_date"`
	CompanyID          int                `json:"company_id"`
	CompanyName        string             `json:"company_name"`
	CustomerSourceID   int                `json:"customer_source_id"`
	Details            string             `json:"details"`
	LossReasonID       interface{}        `json:"loss_reason_id"`
	PipelineID         int                `json:"pipeline_id"`
	PipelineStageID    int                `json:"pipeline_stage_id"`
	PrimaryContactID   interface{}        `json:"primary_contact_id"`
	Priority           string             `json:"priority"`
	Status             string             `json:"status"`
	Tags               []interface{}      `json:"tags"`
	InteractionCount   int                `json:"interaction_count"`
	MonetaryValue      float32            `json:"monetary_value"`
	WinProbability     float32            `json:"win_probability"`
	DateLastContacted  interface{}        `json:"date_last_contacted"`
	LeadsConvertedFrom []interface{}      `json:"leads_converted_from"`
	DateLeadCreated    interface{}        `json:"date_lead_created"`
	DateCreated        int                `json:"date_created"`
	DateModified       int                `json:"date_modified"`
	CustomFields       CopperCustomFields `json:"custom_fields"`
}

type Opportunity struct {
	ID                 int                `json:"id"`
	Name               string             `json:"name"`
	AssigneeID         interface{}        `json:"assignee_id"`
	CloseDate          string             `json:"close_date"`
	CompanyID          int                `json:"company_id"`
	CompanyName        string             `json:"company_name"`
	CustomerSourceID   int                `json:"customer_source_id"`
	Details            string             `json:"details"`
	LossReasonID       interface{}        `json:"loss_reason_id"`
	PipelineID         int                `json:"pipeline_id"`
	PipelineStageID    int                `json:"pipeline_stage_id"`
	PrimaryContactID   interface{}        `json:"primary_contact_id"`
	Priority           string             `json:"priority"`
	Status             string             `json:"status"`
	Tags               []interface{}      `json:"tags"`
	InteractionCount   int                `json:"interaction_count"`
	MonetaryValue      float32            `json:"monetary_value"`
	WinProbability     float32            `json:"win_probability"`
	DateLastContacted  interface{}        `json:"date_last_contacted"`
	LeadsConvertedFrom []interface{}      `json:"leads_converted_from"`
	DateLeadCreated    interface{}        `json:"date_lead_created"`
	DateCreated        int                `json:"date_created"`
	DateModified       int                `json:"date_modified"`
	CustomFields       CopperCustomFields `json:"custom_fields"`
}

func PostCopperRequest(url string, jsonStr string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	var err error = nil
	var response *http.Response = nil
	var req *http.Request = nil

	if err != nil {
		log.Printf("Got error %s", err.Error())
	} else {
		body := strings.NewReader(jsonStr)
		req, err = http.NewRequest("POST", url, body)
		if err != nil {
			log.Printf("Got error %s", err.Error())
		} else {

			header := GetConfigValue("COPPER_API_KEY", "")

			req.Header.Set("X-PW-AccessToken", header)
			header = GetConfigValue("COPPER_USER", "")
			if header == "" {
				return response, errors.New("environment variables empty")
			}
			req.Header.Add("X-PW-UserEmail", header)
			req.Header.Add("X-PW-Application", "developer_api")
			req.Header.Add("Content-Type", "application/json")
			response, err = client.Do(req)
		}
	}
	return response, err
}

func CopperFindPersonByEmailObj(searchtext string) (CopperPerson, error) {
	person := CopperPerson{}
	var err error = nil
	lstxt := strings.ToLower(searchtext)
	if len(lstxt) > 0 {
		urlstr := CP_base_url + CP_people_fragment + "fetch_by_email"
		var r *http.Response
		type email struct {
			Email string `json:"email"`
		}
		search := email{lstxt}
		jsonStr, _ := json.Marshal(search)
		r, err = PostCopperRequest(urlstr, string(jsonStr))
		if err == nil {
			defer r.Body.Close()
			var body []byte
			body, err = io.ReadAll(r.Body)

			if err == nil {
				err = json.Unmarshal(body, &person)
			}
		}
	} else {

		err = errors.New("search text was empty")
	}

	return person, err
}

func CopperListOpportunities(page_number int, pipeline_ids []int, status_ids []int) (Opportunities, error) {
	type funcdata struct {
		PageSize    int    `json:"page_size"`
		SortBy      string `json:"sort_by"`
		PageNumber  int    `json:"page_number"`
		StatusIds   []int  `json:"status_ids"`
		PipelineIds []int  `json:"pipeline_ids,omitempty"`
	}
	opps := Opportunities{}
	var err error = nil

	if page_number == 0 {
		page_number = 1
	}

	if len(status_ids) == 0 {
		status_ids = append(status_ids, 0, 1, 2, 3)
	}

	data := funcdata{100, "name", page_number, status_ids, pipeline_ids}
	url := CP_base_url + CP_opp_fragment + CP_search_fragment
	jsonStr, _ := json.Marshal(data)

	var r *http.Response
	r, err = PostCopperRequest(url, string(jsonStr))
	if err == nil {
		defer r.Body.Close()
		var body []byte
		body, err = io.ReadAll(r.Body)

		if err == nil {
			err = json.Unmarshal(body, &opps)
		}
	}

	return opps, err
}

func CopperGetCustomFieldValue(custom_fields CopperCustomFields, field_id int) interface{} {
	for _, field := range custom_fields {
		if field.CustomFieldDefinitionID == field_id {
			return field.Value
		}
	}

	return nil
}
