package corosync

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"text/template"
)

const corotmpl = `totem {
 version: 2
 cluster_name: {{.Name}}
 secauth: off
 transport: udpu
}

nodelist {{"{"}}{{range $i, $v := .IPs}}
  node {
    ring0_addr: {{$v}}
    nodeid: {{inc $i}}
  }{{end}}
}

quorum {
  provider: corosync_votequorum
}

logging {
  to_logfile: yes
  logfile: /var/log/cluster/corosync.log
  to_syslog: yes
}`

func GenerateConfig(nodeIPs []*net.IP, clusterName string) string {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	t := template.Must(template.New("").Funcs(funcMap).Parse(corotmpl))
	type data struct {
		IPs  []*net.IP
		Name string
	}

	var out bytes.Buffer
	t.Execute(&out, data{IPs: nodeIPs, Name: clusterName})

	return out.String()
}

// GenerateAuthkey calls corosync-keygen in order to generate an authkey and
// write it to the standard location (/etc/corosync/authkey). It returns the
// location of the authkey file and an error, if applicable.
func GenerateAuthkey() (string, error) {
	if err := exec.Command("corosync-keygen", "-l").Run(); err != nil {
		return "", err
	}

	return "/etc/corosync/authkey", nil
}

//Quorum information
//------------------
//Date:             Tue Oct 29 14:11:19 2019
//Quorum provider:  corosync_votequorum
//Nodes:            3
//Node ID:          2
//Ring ID:          1/12
//Quorate:          Yes
//
//Votequorum information
//----------------------
//Expected votes:   3
//Highest expected: 3
//Total votes:      3
//Quorum:           2
//Flags:            Quorate
//
//Membership information
//----------------------
//    Nodeid      Votes    Qdevice Name
//         1          1         NR 192.168.123.11
//         2          1         NR 192.168.123.12 (local)
//         3          1         NR 192.168.123.13
type CorosyncNode struct {
	ID int
	IP net.IP
}

type CorosyncQuorumNode struct {
	CorosyncNode
	Votes int
}

type QuorumStatus struct {
	Nodes         []*CorosyncQuorumNode
	Votes         int
	VotesExpected int
	Quorate       bool
}

var ErrInvalidOutput = errors.New("invalid corosync-quorumtool output")

func getQuorate(output string) (bool, error) {
	r := regexp.MustCompile(`Quorate:\s*(?P<yesno>Yes|No)`)
	match := r.FindStringSubmatch(output)
	if len(match) == 0 {
		return false, fmt.Errorf("error getting quoruate: %s", ErrInvalidOutput.Error())
	}

	return match[1] == "Yes", nil
}

func getVotes(output string) (int, error) {
	r := regexp.MustCompile(`Total votes:\s*(?P<votes>\d+)`)
	match := r.FindStringSubmatch(output)
	if len(match) == 0 {
		return 0, fmt.Errorf("error getting total votes: %s", ErrInvalidOutput.Error())
	}

	return strconv.Atoi(match[1])
}

func getVotesExpected(output string) (int, error) {
	r := regexp.MustCompile(`Expected votes:\s*(?P<votes>\d)`)
	match := r.FindStringSubmatch(output)
	if len(match) == 0 {
		return 0, fmt.Errorf("error getting expected votes: %s", ErrInvalidOutput.Error())
	}

	return strconv.Atoi(match[1])
}

func GetQuorumStatus() (*QuorumStatus, error) {
	cmd := exec.Command("corosync-quorumtool", "-s", "-p", "-i")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// for some dumbass reason corosync-quorumtool returns
			// 1 even if everything is alright... I don't have time
			// for this, just mask out exit code 1.
			if exitError.ExitCode() != 1 {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	output := string(stdoutStderr)

	var status QuorumStatus

	status.Quorate, err = getQuorate(output)
	if err != nil {
		return nil, err
	}

	status.Votes, err = getVotes(output)
	if err != nil {
		return nil, err
	}

	status.VotesExpected, err = getVotesExpected(output)
	if err != nil {
		return nil, err
	}

	// can't be bothered to implement this right now
	status.Nodes = nil

	return &status, nil
}
