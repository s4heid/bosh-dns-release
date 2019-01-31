package command

import (
	"bosh-dns/dns/api"
	"bosh-dns/tlsclient"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type GroupsCmd struct {
	Args               GroupsArgs `positional-args:"true"`
	API                string     `long:"api" env:"DNS_API_ADDRESS" description:"API address to talk to"`
	TLSCACertPath      string     `long:"ca-cert-path" env:"DNS_API_TLS_CA_CERT_PATH" description:"CA certificate to use for mutual LS"`
	TLSCertificatePath string     `long:"certificate-path" env:"DNS_API_TLS_CERTIFICATE_PATH" description:"Client certificate to use for mutual LS"`
	TLSPrivateKeyPath  string     `long:"private-key-path" env:"DNS_API_TLS_PRIVATE_KEY_PATH" description:"Client key to use for mutual LS"`

	UI ui.UI
}

type GroupsArgs struct {
	Query string `positional-arg-name:"QUERY" description:"BOSH-DNS query formatted instance filter"`
}

func (o *GroupsCmd) Execute(args []string) error {
	logger := boshlog.NewLogger(boshlog.LevelNone)
	if o.UI == nil {
		confUI := ui.NewConfUI(logger)
		confUI.EnableColor()
		o.UI = confUI
	}

	client, err := tlsclient.NewFromFiles("api.bosh-dns", o.TLSCACertPath, o.TLSCertificatePath, o.TLSPrivateKeyPath, logger)
	if err != nil {
		return err
	}

	requestURL := o.API + "/groups"

	response, err := client.Get(requestURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to retrieve groups: Got %s", response.Status)
	}

	table := boshtbl.Table{
		FillFirstColumn: true,
		Header: []boshtbl.Header{
			boshtbl.NewHeader("JobName"),
			boshtbl.NewHeader("LinkName"),
			boshtbl.NewHeader("LinkType"),
			boshtbl.NewHeader("GroupID"),
			boshtbl.NewHeader("HealthState"),
		},
	}

	decoder := json.NewDecoder(response.Body)

	for decoder.More() {
		var jsonRow api.Group

		err := decoder.Decode(&jsonRow)
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(jsonRow.JobName),
			boshtbl.NewValueString(jsonRow.LinkName),
			boshtbl.NewValueString(jsonRow.LinkType),
			boshtbl.NewValueInt(jsonRow.GroupID),
			boshtbl.NewValueString(jsonRow.HealthState),
		})
	}

	o.UI.PrintTable(table)

	return nil
}