package main

import (
	"fmt"
	"github.com/dynport/gocli"
	"github.com/dynport/gocloud/jiffybox"
	"os"
	"strings"
)

func init() {
	router.RegisterFunc("jb/backups/list", JiffyBoxListBackups, "List all running boxes")
	router.Register("jb/backups/create", &JiffyBoxCreateBackup{}, "Create manual backup from box")
}

const USAGE_CREATE_BACKUP = "id"

type JiffyBoxCreateBackup struct {
	Id int `cli:"type=arg required=true"`
}

func (a *JiffyBoxCreateBackup) Run() error {
	logger.Infof("creating backup for box %d", a.Id)
	if e := client().CreateBackup(a.Id); e != nil {
		return e
	}
	logger.Infof("created backup for box %d", a.Id)
	return nil
}

func init() {
	router.Register("jb/servers/shutdown", &JiffyBoxStopServer{}, "Shutdown Server")
}

type JiffyBoxStopServer struct {
	Id int `cli:"type=arg required=true"`
}

func (a *JiffyBoxStopServer) Run() error {
	s, e := client().ShutdownServer(a.Id)
	if e != nil {
		return e
	}
	logger.Infof("stopped server %d", a.Id)
	printServer(s)
	return nil
}

const USAGE_FREEZE_SERVER = "id"

func init() {
	router.Register("jb/servers/freeze", &JiffyBoxFreezeServer{}, "Freeze Server")
}

type JiffyBoxFreezeServer struct {
	Id int `cli:"type=arg required=true"`
}

func (a *JiffyBoxFreezeServer) Run() error {
	s, e := client().JiffyBox(a.Id)
	if e != nil {
		return e
	}
	if s.Running {
		return fmt.Errorf("Server must not be running!")
	}
	s, e = client().FreezeServer(a.Id)
	if e != nil {
		return e
	}
	logger.Infof("froze server %d", a.Id)
	printServer(s)
	return nil
}

func init() {
	router.Register("jb/servers/start", &JiffyBoxStartServer{}, "Start Server")
}

type JiffyBoxStartServer struct {
	PlanId int `cli:"type=opt required=true short=p"`
	BoxId  int `cli:"type=arg required=true"`
}

func (a *JiffyBoxStartServer) Run() error {
	s, e := client().StartServer(a.BoxId, a.PlanId)
	if e != nil {
		return e
	}
	logger.Infof("started server %d", a.BoxId)
	printServer(s)
	return nil
}

func init() {
	router.Register("jb/servers/thaw", &JiffyBoxThawServer{}, "Thaw Server")
}

type JiffyBoxThawServer struct {
	PlanId int `cli:"type=opt short=p required=true"`
	BoxId  int `cli:"type=arg required=true"`
}

func (a *JiffyBoxThawServer) Run() error {
	s, e := client().ThawServer(a.BoxId, a.PlanId)
	if e != nil {
		return e
	}
	logger.Infof("thawed server %d", a.BoxId)
	printServer(s)
	return nil
}

func init() {
	router.RegisterFunc("jb/servers/list", JiffyBoxListServersAction, "List Servers")
	router.Register("jb/servers/show", &JiffyBoxShowServersAction{}, "Show Server")
	router.Register("jb/servers/clone", &JiffyBoxCloneServer{}, "Clone Server")
}

type JiffyBoxCloneServer struct {
	BoxId  int    `cli:"type=arg required=true"`
	Name   string `cli:"type=arg required=true"`
	PlanId int    `cli:"type=opt short=p"`
}

func (a *JiffyBoxCloneServer) Run() error {
	s, e := client().JiffyBox(a.BoxId)
	if e != nil {
		return e
	}
	if s.Frozen() {
		return fmt.Errorf("Server must not be frozen!")
	}
	opts := &jiffybox.CreateOptions{
		PlanId:   a.PlanId,
		Name:     a.Name,
		Password: os.Getenv("JIFFYBOX_DEFAULT_PASSWORD"),
	}
	logger.Infof("cloning server %d with %#v", a.BoxId, opts)
	s, e = client().CloneServer(a.BoxId, opts)
	if e != nil {
		return e
	}
	logger.Infof("cloned server %d", a.BoxId)
	printServer(s)
	return nil
}

type JiffyBoxShowServersAction struct {
	BoxId int `cli:"type=arg required=true"`
}

func (a *JiffyBoxShowServersAction) Run() error {
	server, e := client().JiffyBox(a.BoxId)
	if e != nil {
		return e
	}
	printServer(server)
	return nil
}

func printServer(server *jiffybox.Server) {
	table := gocli.NewTable()
	table.Add("Id", server.Id)
	table.Add("Name", server.Name)
	table.Add("Status", server.Status)
	table.Add("Created", server.CreatedAt().Format(TIME_FORMAT))
	table.Add("Host", server.Host)
	table.Add("Running", server.Running)
	table.Add("RecoverymodeActive", server.RecoverymodeActive)
	table.Add("Plan", server.Plan.Id)
	table.Add("Cpu", server.Plan.Cpus)
	table.Add("RAM", server.Plan.RamInMB)
	table.Add("IsBeingCopied", server.IsBeingCopied)
	table.Add("ManualBackupRunning", server.ManualBackupRunning)
	if server.ActiveProfile != nil {
		table.Add("Profile Name", server.ActiveProfile.Name)
		table.Add("Profile Kernel", server.ActiveProfile.Kernel)
	}
	i := 0
	for k, v := range server.Ips {
		key := ""
		if i == 0 {
			key = "Ips"
			i++
		}
		table.Add(key, k+": "+strings.Join(v, ", "))
	}
	fmt.Println(table)
}

func init() {
	router.RegisterFunc("jb/plans/list", JiffyBoxListPlansAction, "List Plans")
	router.RegisterFunc("jb/distributions/list", JiffyBoxListDistributionsAction, "List Distributions")
	router.Register("jb/servers/delete", &JiffyBoxDeleteAction{}, "Delete Jiffybox")
}

func JiffyBoxListBackups() error {
	backups, e := client().Backups()
	if e != nil {
		return e
	}

	table := gocli.NewTable()
	for _, backup := range backups {
		table.Add(backup.Id, backup.ServerId, backup.Key, backup.CreatedAt().Format(TIME_FORMAT))
	}
	fmt.Println(table)
	return nil
}

const USAGE_DELETE = "id"

type JiffyBoxDeleteAction struct {
	BoxId int `cli:"type=arg required=true"`
}

func (a *JiffyBoxDeleteAction) Run() error {
	logger.Infof("deleting box with id %s", a.BoxId)
	e := client().DeleteJiffyBox(a.BoxId)
	if e != nil {
		return e
	}
	logger.Info("deleted box")
	return nil
}

func client() *jiffybox.Client {
	return jiffybox.NewFromEnv()
}

func JiffyBoxListDistributionsAction() error {
	distributions, e := client().Distributions()
	if e != nil {
		return e
	}
	table := gocli.NewTable()
	table.Add("Key", "Name", "Min Disk Size", "Default Kernel")
	for _, distribution := range distributions {
		table.Add(distribution.Key, distribution.Name, distribution.MinDiskSizeMB, distribution.DefaultKernel)
	}
	fmt.Println(table)
	return nil
}

const (
	CLI_NAME             = "--name"
	CLI_PLAN_ID          = "--plan-id"
	CLI_DISTRIBUTION     = "--distribution"
	DEFAULT_PLAN_ID      = 20
	DEFAULT_DISTRIBUTION = "ubuntu_12_4_lts_64bit"
	USAGE_CLONE_SERVER   = "id"
	HOURS_PER_MONTH      = 365 * 24.0 / 12.0
)

func init() {
	args := gocli.NewArgs(nil)
	args.RegisterString(CLI_NAME, "name", false, "", "Name of the new box")
	args.RegisterInt(CLI_PLAN_ID, "plan_id", false, DEFAULT_PLAN_ID, "Plan id")
	args.RegisterString(CLI_DISTRIBUTION, "distribution", false, DEFAULT_DISTRIBUTION, "Distribution")
	router.Register("jb/servers/create", &JiffyBoxCreateAction{}, "Create new JiffyBox")
}

func JiffyBoxListPlansAction() error {
	plans, e := client().Plans()
	if e != nil {
		return e
	}
	table := gocli.NewTable()
	table.Add("Id", "Name", "Cpu", "Ram", "Disk", "Price/Hour", "Price/Month")
	for _, plan := range plans {
		table.Add(
			plan.Id, plan.Name, plan.Cpus, plan.RamInMB, plan.DiskSizeInMB,
			fmt.Sprintf("%.02f €", plan.PricePerHour),
			fmt.Sprintf("%.2f €", plan.PricePerHour*HOURS_PER_MONTH),
		)
	}
	fmt.Println(table)
	return nil
}

type JiffyBoxCreateAction struct {
	Name         string `cli:"type=arg required=true"`
	PlanId       int    `cli:"type=opt short=p required=true"`
	Distribution string `cli:"type=opt short=d required=true"`
}

func (a *JiffyBoxCreateAction) Run() error {
	logger.Infof("creating new jiffybox")
	opts := &jiffybox.CreateOptions{
		Name:         a.Name,
		PlanId:       a.PlanId,
		Distribution: a.Distribution,
		UseSshKey:    true,
		Password:     os.Getenv("JIFFYBOX_DEFAULT_PASSWORD"),
	}
	s, e := client().CreateJiffyBox(opts)
	if e != nil {
		return e
	}
	fmt.Println("created server!")
	printServer(s)
	return nil
}

const TIME_FORMAT = "2006-01-02T15:04:05"

func JiffyBoxListServersAction() error {
	servers, e := client().JiffyBoxes()
	if e != nil {
		return e
	}
	if len(servers) == 0 {
		fmt.Println("no boxes found")
		return nil
	}
	table := gocli.NewTable()
	table.Add("Created", "Id", "Status", "Running", "Name", "Cpu", "RAM", "Ip")
	for _, server := range servers {
		table.Add(server.CreatedAt().Format(TIME_FORMAT), server.Id, server.Status, server.Running, server.Name, server.Plan.Cpus, server.Plan.RamInMB, server.PublicIp())
	}
	fmt.Println(table)
	return nil
}
