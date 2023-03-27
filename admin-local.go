package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v48/github"
	"github.com/owasp-foundation/admin-local-go/shared"
	"github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/customer"
	"golang.org/x/oauth2"
)

func get_address_as_string(address shared.CopperAddress) string {
	var address_string string = ""
	if address.Street != "" {
		address_string += address.Street
		address_string += "\n"
	}
	if address.City != "" {
		address_string += address.City
		address_string += "\n"
	}
	if address.State != "" {
		address_string += address.State
		address_string += "\n"
	}
	if address.Country != "" {
		address_string += address.Country
		address_string += "\n"
	}
	if address.PostalCode != "" {
		address_string += address.PostalCode
	}

	return address_string
}

func get_copper_person(email string) shared.CopperPerson {
	cp, _ := shared.CopperFindPersonByEmailObj(email)

	return cp
}

func fill_member_row_data(row []string, person shared.CopperPerson, customer *stripe.Customer, metadata map[string]string) []string {
	firstName := strings.TrimSpace(person.FirstName)
	lastName := strings.TrimSpace(person.LastName)
	if firstName == "" { // no first name, get both first and last from Stripe Customer
		names := strings.Split(customer.Name, " ")
		if len(names) >= 1 {
			for i, v := range names {
				if i == 0 {
					firstName = v
				} else {
					lastName = v + " "
				}
			}
			lastName = strings.TrimRight(lastName, " ")
		}
	} else if lastName == "" {
		names := strings.Split(customer.Name, " ")
		if len(names) >= 1 {
			for i, v := range names {
				if i > 0 {
					lastName = v + " "
				}
			}
			lastName = strings.TrimRight(lastName, " ")
		}
	}

	row = append(row, firstName)
	row = append(row, lastName)
	emailstr := ""
	owasp_email := ""
	for _, email := range person.Emails {
		if strings.Contains(email.Email, "@owasp.org") {
			owasp_email = email.Email
			emailstr = owasp_email
			break
		}
	}

	if owasp_email == "" {
		meta_email := metadata["owasp_email"]
		if strings.TrimSpace(meta_email) != "" {
			owasp_email = strings.TrimSpace(meta_email)
			emailstr = owasp_email
		}
	}

	for _, email := range person.Emails {
		if email.Email == owasp_email {
			continue
		} else if emailstr != "" {
			emailstr += "\n"
		}

		emailstr += email.Email
	}
	if !strings.Contains(emailstr, customer.Email) {
		if emailstr != "" {
			emailstr += "\n"
		}
		emailstr += customer.Email
	}
	row = append(row, emailstr)
	phonestr := ""
	for _, phone := range person.PhoneNumbers {
		if phonestr != "" {
			phonestr += "\n"
		}
		phonestr += phone.Number
	}
	row = append(row, phonestr)
	row = append(row, person.Address.Street)
	row = append(row, person.Address.City)
	row = append(row, person.Address.State)
	row = append(row, person.Address.Country)
	row = append(row, person.Address.PostalCode)
	row = append(row, metadata["membership_type"])
	row = append(row, metadata["membership_start"])
	row = append(row, metadata["membership_end"])
	row = append(row, metadata["membership_recurring"])

	github := shared.CopperGetCustomFieldValue(person.CustomFields, shared.CP_person_github_username)
	row = append(row, fmt.Sprintf("%v", github))

	tagstr := ""
	for _, tag := range person.Tags {
		tagstr += tag
		tagstr += "\n"
	}

	row = append(row, tagstr)

	return row
}

func export_members_for_ym() {
	fmt.Println("Exporting members to csv")

	member_data := shared.MemberData{}

	skey := shared.GetConfigValue("STRIPE_SECRET", "")

	stripe.Key = skey
	params := &stripe.CustomerSearchParams{}
	params.Query = *stripe.String("-metadata['membership_type']:null")

	iter := customer.Search(params)
	expiry := time.Now().AddDate(0, 0, -1)
	records := [][]string{
		{"first_name", "last_name", "emails", "phone_numbers", "street_address", "city", "state", "country", "postal_code", "membership_type", "membership_start", "membership_end", "membership_recurring", "github_id", "tags"},
	}

	for iter.Next() {
		row := make([]string, 0)
		current := iter.Customer()
		metadata := current.Metadata
		var copper_person shared.CopperPerson
		if metadata != nil {
			member_type := strings.Trim(strings.ToLower(metadata["membership_type"]), " ")

			if strings.Contains(member_type, "lifetime") {
				member_data.Lifetime++
				copper_person = get_copper_person(current.Email)
				row = fill_member_row_data(row, copper_person, current, metadata)
			} else {
				member_end := metadata["membership_end"]
				end_date, _ := shared.StringToDateTimeHelper(member_end)
				if end_date.After(expiry) {
					if strings.Contains(member_type, "one") {
						member_data.One++
						copper_person = get_copper_person(current.Email)
						row = fill_member_row_data(row, copper_person, current, metadata)
					} else if strings.Contains(member_type, "two") {
						member_data.Two++
						copper_person = get_copper_person(current.Email)
						row = fill_member_row_data(row, copper_person, current, metadata)
					} else if strings.Contains(member_type, "complimentary") {
						member_data.Complimentary++
						copper_person = get_copper_person(current.Email)
						row = fill_member_row_data(row, copper_person, current, metadata)
					}
				}
			}
		}
		if len(row) > 0 {
			records = append(records, row)
		}
	}

	filename := fmt.Sprintf("members_%s.csv", strings.ReplaceAll(time.Now().String(), " ", "_"))
	csv_file, ferr := os.Create(filename)

	if ferr != nil {
		fmt.Println("Failed to open file")
	} else {
		defer csv_file.Close()
		w := csv.NewWriter(csv_file)
		w.WriteAll(records) // calls Flush internally
	}
	fmt.Println("Done")
}

func get_repos_matching(ctx context.Context, client *github.Client, match string) []*github.Repository {
	// get all pages of results

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
		Type:        "public",
	}

	var allRepos []*github.Repository
	var retRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, "owasp", opt)
		if err != nil {
			fmt.Println(err.Error())
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for _, repo := range allRepos {
		if strings.Contains(repo.GetName(), match) && *repo.HasPages {
			retRepos = append(retRepos, repo)
		}
	}

	return retRepos
}

type owasp_project struct {
	Name          string
	ProjectType   string
	Level         string
	Repo          string
	Website       string
	Updated       time.Time
	CodeUrl       string
	LastCommit    string
	IssueCount    int
	ExternalLinks string
}

func (p *owasp_project) initialize(indexReader io.ReadCloser, infoReader io.ReadCloser, tabReaders []io.ReadCloser, repo *github.Repository) {
	scanner := bufio.NewScanner(indexReader)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, "title:") {
			_, title, _ := strings.Cut(txt, ":")
			p.Name = strings.TrimSpace(strings.ToLower(title))
		} else if strings.Contains(txt, "type:") {
			_, ptype, _ := strings.Cut(txt, ":")
			p.ProjectType = strings.TrimSpace(strings.ToLower(ptype))
		} else if strings.Contains(txt, "level:") {
			_, level, _ := strings.Cut(txt, ":")
			level = strings.TrimSpace(strings.ToLower(level))
			switch level {
			case "1":
				p.Level = ""
			case "2":
				p.Level = "Incubator"
			case "3":
				p.Level = "Lab"
			case "3.5":
				p.Level = "Production"
			case "4":
				p.Level = "Flagship"
			}
		} else if strings.Contains(txt, "https://") && !strings.Contains(txt, "https://owasp.org") && !strings.Contains(txt, "https://github.com") {
			//likely an external link...
			scantxt := txt
			scantxt = strings.ReplaceAll(scantxt, "]", " ")
			scantxt = strings.ReplaceAll(scantxt, ")", " ")
			scantxt = strings.ReplaceAll(scantxt, ",", " ")
			linkStart := strings.Index(scantxt, "https://")
			for linkStart > -1 {
				linkEnd := strings.Index(scantxt[linkStart:], " ")
				if linkEnd == -1 {
					linkEnd = len(scantxt) - 1
				} else {
					linkEnd += linkStart
				}
				if !strings.Contains(scantxt[linkStart:linkEnd], "owasp.org") && !strings.Contains(scantxt[linkStart:linkEnd], "github.com") {
					if !strings.Contains(p.ExternalLinks, scantxt[linkStart:linkEnd]) {
						p.ExternalLinks += scantxt[linkStart:linkEnd] + "\n"
					}
				}
				scantxt = scantxt[linkEnd:]
				linkStart = strings.Index(scantxt, "https://")
			}
		} else if strings.Contains(txt, "https://github.com") {
			//possibly the code link, let's add it
			scantxt := txt
			scantxt = strings.ReplaceAll(scantxt, "]", " ")
			scantxt = strings.ReplaceAll(scantxt, ")", " ")
			scantxt = strings.ReplaceAll(scantxt, ",", " ")
			linkStart := strings.Index(scantxt, "https://github.com")
			for linkStart > -1 {
				linkEnd := strings.Index(scantxt[linkStart:], " ")
				if linkEnd == -1 {
					linkEnd = len(scantxt) - 1
				} else {
					linkEnd += linkStart
				}

				codeUrl := scantxt[linkStart:linkEnd]
				owner, coderepo := get_github_components(codeUrl)
				if owner != "" && coderepo != "" {
					codeUrl = "https://github.com/"
					codeUrl += owner + "/"
					codeUrl += coderepo
				}
				if codeUrl != "" && !strings.Contains(p.CodeUrl, codeUrl) {

					p.CodeUrl += codeUrl + "\n"
				}

				scantxt = scantxt[linkEnd:]
				linkStart = strings.Index(scantxt, "https://github.com")
			}
		}
	}
	scanner = bufio.NewScanner(infoReader)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.Contains(txt, "https://") && !strings.Contains(txt, "https://owasp.org") && !strings.Contains(txt, "https://github.com") {
			//likely an external link...
			scantxt := txt
			scantxt = strings.ReplaceAll(scantxt, "]", " ")
			scantxt = strings.ReplaceAll(scantxt, ")", " ")
			scantxt = strings.ReplaceAll(scantxt, ",", " ")
			linkStart := strings.Index(scantxt, "https://")
			for linkStart > -1 {
				linkEnd := strings.Index(scantxt[linkStart:], " ")
				if linkEnd == -1 {
					linkEnd = len(scantxt) - 1
				} else {
					linkEnd += linkStart
				}
				if !strings.Contains(scantxt[linkStart:linkEnd], "owasp.org") && !strings.Contains(scantxt[linkStart:linkEnd], "github.com") {
					if !strings.Contains(p.ExternalLinks, scantxt[linkStart:linkEnd]) {
						p.ExternalLinks += scantxt[linkStart:linkEnd] + "\n"
					}
				}
				scantxt = scantxt[linkEnd:]
				linkStart = strings.Index(scantxt, "https://")
			}
		} else if strings.Contains(txt, "https://github.com") {
			//possibly the code link, let's add it
			scantxt := txt
			scantxt = strings.ReplaceAll(scantxt, "]", " ")
			scantxt = strings.ReplaceAll(scantxt, ")", " ")
			scantxt = strings.ReplaceAll(scantxt, ",", " ")
			linkStart := strings.Index(scantxt, "https://github.com")
			for linkStart > -1 {
				linkEnd := strings.Index(scantxt[linkStart:], " ")
				if linkEnd == -1 {
					linkEnd = len(scantxt) - 1
				} else {
					linkEnd += linkStart
				}

				codeUrl := scantxt[linkStart:linkEnd]
				owner, coderepo := get_github_components(codeUrl)
				if owner != "" && coderepo != "" {
					codeUrl = "https://github.com/"
					codeUrl += owner + "/"
					codeUrl += coderepo
				}
				if codeUrl != "" && !strings.Contains(p.CodeUrl, codeUrl) {

					p.CodeUrl += codeUrl + "\n"
				}

				scantxt = scantxt[linkEnd:]
				linkStart = strings.Index(scantxt, "https://github.com")
			}
		}
	}

	for _, reader := range tabReaders {
		scanner = bufio.NewScanner(reader)
		for scanner.Scan() {
			txt := scanner.Text()
			if strings.Contains(txt, "https://") && !strings.Contains(txt, "https://owasp.org") && !strings.Contains(txt, "https://github.com") {
				//likely an external link...
				scantxt := txt
				scantxt = strings.ReplaceAll(scantxt, "]", " ")
				scantxt = strings.ReplaceAll(scantxt, ")", " ")
				scantxt = strings.ReplaceAll(scantxt, ",", " ")
				linkStart := strings.Index(scantxt, "https://")
				for linkStart > -1 {
					linkEnd := strings.Index(scantxt[linkStart:], " ")
					if linkEnd == -1 {
						linkEnd = len(scantxt) - 1
					} else {
						linkEnd += linkStart
					}
					if !strings.Contains(scantxt[linkStart:linkEnd], "owasp.org") && !strings.Contains(scantxt[linkStart:linkEnd], "github.com") {
						if !strings.Contains(p.ExternalLinks, scantxt[linkStart:linkEnd]) {
							p.ExternalLinks += scantxt[linkStart:linkEnd] + "\n"
						}
					}
					scantxt = scantxt[linkEnd:]
					linkStart = strings.Index(scantxt, "https://")
				}
			} else if strings.Contains(txt, "https://github.com") {
				//possibly the code link, let's add it
				scantxt := txt
				scantxt = strings.ReplaceAll(scantxt, "]", " ")
				scantxt = strings.ReplaceAll(scantxt, ")", " ")
				scantxt = strings.ReplaceAll(scantxt, ",", " ")
				linkStart := strings.Index(scantxt, "https://github.com")
				for linkStart > -1 {
					linkEnd := strings.Index(scantxt[linkStart:], " ")
					if linkEnd == -1 {
						linkEnd = len(scantxt) - 1
					} else {
						linkEnd += linkStart
					}

					codeUrl := scantxt[linkStart:linkEnd]
					owner, coderepo := get_github_components(codeUrl)
					if owner != "" && coderepo != "" {
						codeUrl = "https://github.com/"
						codeUrl += owner + "/"
						codeUrl += coderepo
					}
					if codeUrl != "" && !strings.Contains(p.CodeUrl, codeUrl) {

						p.CodeUrl += codeUrl + "\n"
					}

					scantxt = scantxt[linkEnd:]
					linkStart = strings.Index(scantxt, "https://github.com")
				}
			}
		}
	}
	p.Updated = repo.GetUpdatedAt().Time
}

func get_github_components(cr string) (string, string) {
	owner := ""
	repo := ""

	// format for github url is https://github.com/owner/repo
	parts := strings.Split(cr, "/")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "https:" && part != "github.com" && part != "" {
			if owner == "" {
				owner = part
			} else if repo == "" {
				repo = part
			} else {
				break
			}
		}
	}

	return owner, repo
}

func fill_project_row_data(client *github.Client, ctx context.Context, repo *github.Repository) []string {
	//{"Name", "Level", "Type", "Repo", "Website URL", "Website Updated", "Code URL", "Last Commit", "Open Issue Count", "External Links"},
	row := make([]string, 0)
	// need to get the index.md file and the leaders.md file and any and all tab_xxx.md files
	indexReader, _, err := client.Repositories.DownloadContents(ctx, "owasp", repo.GetName(), "index.md", nil)
	if err != nil {
		fmt.Println("failure to get index.md on" + repo.GetName() + " with error " + err.Error())
		return row
	}
	defer indexReader.Close()

	infoReader, _, err := client.Repositories.DownloadContents(ctx, "owasp", repo.GetName(), "info.md", nil)
	if err != nil {
		fmt.Println("failure to get info.md on" + repo.GetName() + " with error " + err.Error())
	}
	defer infoReader.Close()

	var tabReaders []io.ReadCloser = make([]io.ReadCloser, 0)

	_, dirContent, _, err := client.Repositories.GetContents(ctx, "owasp", repo.GetName(), "/", nil)
	if err != nil {
		panic(err)
	} else {
		for _, content := range dirContent {
			if strings.Contains(content.GetName(), "tab_") {
				tReader, _, err := client.Repositories.DownloadContents(ctx, "owasp", repo.GetName(), content.GetName(), nil)
				if err != nil {
					fmt.Println("failed to get contents of " + repo.GetName() + " with error " + err.Error())
					return row
				}
				defer tReader.Close()
				tabReaders = append(tabReaders, tReader)
			}
		}
	}

	var p owasp_project
	p.initialize(indexReader, infoReader, tabReaders, repo)

	if p.CodeUrl != "" {
		coderepos := strings.Split(p.CodeUrl, "\n")
		for _, cr := range coderepos {
			owner, coderepo := get_github_components(cr)
			if owner == "" || coderepo == "" {
				continue
			}
			ws, resp, err := client.Repositories.ListCommitActivity(ctx, owner, coderepo)
			if resp.StatusCode == 202 {
				time.Sleep(time.Second * 5)
				ws, _, err = client.Repositories.ListCommitActivity(ctx, owner, coderepo)
			}
			if err != nil {
				fmt.Println("Could not get commit activity for " + owner + "/" + coderepo + " with error " + err.Error())
			} else {
				if len(ws) > 0 {
					p.LastCommit = ws[0].String() + "\n"
				}
				//need to get the repo and see issue count...
				coderepo, _, err := client.Repositories.Get(ctx, owner, coderepo)
				if err == nil {
					p.IssueCount += *coderepo.OpenIssuesCount
				} else {
					fmt.Println("Could not get issue count for " + cr + " with error " + err.Error())
				}
			}
		}
	}
	row = append(row, p.Name)
	row = append(row, p.Level)
	row = append(row, p.ProjectType)
	row = append(row, repo.GetName())
	row = append(row, "https://owasp.org/"+repo.GetName())
	row = append(row, p.Updated.String())
	row = append(row, p.CodeUrl)
	row = append(row, p.LastCommit)
	row = append(row, fmt.Sprintf("%v", p.IssueCount))
	row = append(row, p.ExternalLinks)
	return row
}

func project_audit() {
	fmt.Println("Performing audit...")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: shared.GetConfigValue("GH_APITOKEN", "")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repos := get_repos_matching(ctx, client, "www-project-")
	records := [][]string{
		{"Name", "Level", "Type", "Repo", "Website URL", "Website Updated", "Code URL", "Last Commit", "Open Issue Count", "External Links"},
	}
	for _, repo := range repos {
		row := fill_project_row_data(client, ctx, repo)
		if len(row) > 0 {
			records = append(records, row)
		}
	}
	filename := fmt.Sprintf("projects_%s.csv", strings.ReplaceAll(time.Now().String(), " ", "_"))
	csv_file, ferr := os.Create(filename)

	if ferr != nil {
		fmt.Println("Failed to open file")
	} else {
		defer csv_file.Close()
		w := csv.NewWriter(csv_file)
		w.WriteAll(records) // calls Flush internally
	}
	fmt.Println("Done")
}

// functions in this quick and dirty admin tool:
//
// # export_members_for_ym()
// # exports members from Stripe/Copper to a csv file to be imported into YourMembership
//
// # project_audit
// # prepares a file which indicates project name, leaders, last website update, last commit, last issue, external links on website
func main() {
	project_audit()
	//export_members_for_ym()
}
