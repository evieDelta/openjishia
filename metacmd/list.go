package metacmd

import (
	"sort"
	"strings"

	"codeberg.org/eviedelta/drc"
	"codeberg.org/eviedelta/trit"
	"github.com/bwmarrin/discordgo"
)

// cList is a command that lists all commands
var cList = &drc.Command{
	Name:         "commands",
	Manual:       []string{"gives a list of all public commands", "use -abc to sort it alphabetically and -zyx to sort it reverse alphabetically"},
	CommandPerms: discordgo.PermissionSendMessages,
	Permissions: drc.Permissions{
		BotAdmin: trit.False,
	},
	Config: drc.CfgCommand{
		Listable: true,
		ReactOn:  drc.ActOn{
			//			Success: trit.True,
		},
		MinimumArgs: 0,
		BoolFlags: map[string]bool{
			"all": false,
			"abc": true,
			"zyx": false,
		},
	},
	Exec: fList,
}

func fList(ctx *drc.Context) error {
	all := ctx.BoolFlags["all"]
	if !ctx.Han.IsUserBotAdmin(ctx.Mes.Author.ID) {
		all = false
	}

	var list []drc.CommandListStructure
	if len(ctx.RawArgs) < 1 {
		list = ctx.Han.Commands.ListStructured(all, 2)
	} else {
		command, _ := ctx.Han.FetchCommand(ctx.Han.Commands, drc.FetcherData{Args: ctx.RawArgs})
		list = command.Subcommands.ListStructured(all, 2)
	}

	user, err := ctx.Ses.State.UserChannelPermissions(ctx.Mes.Author.ID, ctx.Mes.ChannelID)
	if err != nil {
		return err
	}

	ls := subList(list, 0, ctx.BoolFlags["abc"], ctx.BoolFlags["zyx"], drc.Permissions{Discord: user, BotAdmin: func(u string) trit.Trit { t := new(trit.Trit); t.SetIfUnset(ctx.Han.IsUserBotAdmin(u)); return *t }(ctx.Mes.Author.ID)})
	return ctx.DumpReply("", ls)
}

type sortList []drc.CommandListStructure

func (a sortList) Len() int           { return len(a) }
func (a sortList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortList) Less(i, j int) bool { return a[i].Name < a[j].Name }

func subList(cl []drc.CommandListStructure, depth int, abc, zyx bool, uperm drc.Permissions) string {
	if len(cl) < 1 {
		return ""
	}
	sl := sortList(cl)
	if zyx {
		sort.Stable(sort.Reverse(sl))
	} else if abc {
		sort.Stable(sl)
	}

	//	fmt.Println("user is bot admin? ", uperm.BotAdmin.Bool())

	ls := ""
	for _, x := range sl {
		if x.Name == "help" || x.Name == "manual" {
			continue
		}
		if x.Permissions.Discord&uperm.Discord != x.Permissions.Discord || x.Permissions.DiscordChannel&uperm.Discord != x.Permissions.DiscordChannel || (x.Permissions.BotAdmin.Bool() && !uperm.BotAdmin.Bool()) {
			continue
		}
		ls += strings.Repeat(" -", depth)
		ls += " : " + x.Name + "\n"
		ls += subList(x.SubLists, depth+1, abc, zyx, uperm)
	}
	return ls
}
