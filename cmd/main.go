package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tcooper-uk/go-todo/internal"
	"github.com/tcooper-uk/go-todo/internal/storage"
	s "github.com/tcooper-uk/go-todo/internal/storage"
	"github.com/tcooper-uk/go-todo/internal/storage/db"
)

// tagList implements flag.Value for repeatable --tag flags.
type tagList []string

func (t *tagList) String() string { return strings.Join(*t, ",") }
func (t *tagList) Set(v string) error {
	*t = append(*t, v)
	return nil
}

func main() {
	args := os.Args[1:]

	// Global --backend flag parsed before the subcommand.
	globalFlags := flag.NewFlagSet("", flag.ContinueOnError)
	backendFlag := globalFlags.String("backend", "", "backend: sqlite|file|cloud (overrides $TODO_BACKEND)")
	globalFlags.Parse(args)
	remainingArgs := globalFlags.Args()

	mode := resolveMode(*backendFlag)

	filePath, err := storage.Setup(mode)
	exitOnErr(err)

	var store s.TodoStore

	switch mode {
	case storage.DbMode:
		store, err = db.NewSQLLiteStorage(filePath)
	case storage.FileMode:
		store = s.NewLocalFileStore(filePath)
	case storage.CloudMode:
		store, err = db.NewCloudStore(&db.CloudStoreConfig{
			ProjectId: db.ProjectId,
			KeyFile:   filePath,
		})
	}
	exitOnErr(err)

	if len(remainingArgs) == 0 {
		printItems(store, s.ListOptions{})
		return
	}

	// Single numeric arg: show that item.
	if id, err := strconv.ParseInt(remainingArgs[0], 0, 0); len(remainingArgs) == 1 && err == nil {
		item := store.GetItem(int(id))
		if item == nil {
			fmt.Printf("Cannot find item with ID %d\n", id)
			os.Exit(1)
		}
		printItem(item)
		return
	}

	cmd := remainingArgs[0]
	cmdArgs := remainingArgs[1:]

	switch cmd {
	case "list", "l", "ps", "ls":
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		all := fs.Bool("all", false, "show done items too")
		onlyDone := fs.Bool("done", false, "show only done items")
		priority := fs.String("priority", "", "filter by priority: low|medium|high")
		var tags tagList
		fs.Var(&tags, "tag", "filter by tag (repeatable)")
		overdue := fs.Bool("overdue", false, "show only overdue items")
		fs.Parse(cmdArgs)

		opts := s.ListOptions{
			ShowDone: *all,
			OnlyDone: *onlyDone,
			Priority: *priority,
			Overdue:  *overdue,
		}
		if len(tags) > 0 {
			opts.Tag = tags[0]
		}
		printItems(store, opts)

	case "delete", "remove", "d", "rm":
		ids := parseIds(cmdArgs...)
		store.DeleteItem(ids...)

	case "add", "create", "put", "a":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		priority := fs.String("priority", "", "priority: low|medium|high")
		due := fs.String("due", "", "due date: YYYY-MM-DD")
		var tags tagList
		fs.Var(&tags, "tag", "tag (repeatable)")
		fs.Parse(cmdArgs)

		name := strings.Join(fs.Args(), " ")
		if name == "" {
			var editorErr error
			name, editorErr = openInEditor("")
			exitOnErr(editorErr)
		}
		if name == "" {
			fmt.Println("No content provided.")
			os.Exit(1)
		}

		todo := internal.Todo{
			Name:     name,
			Priority: internal.Priority(*priority),
			Tags:     []string(tags),
		}
		if *due != "" {
			t, err := time.Parse("2006-01-02", *due)
			if err != nil {
				fmt.Printf("Invalid date %q — use YYYY-MM-DD\n", *due)
				os.Exit(1)
			}
			todo.DueDate = &t
		}
		store.AddItem(todo)

	case "e", "edit", "update":
		fs := flag.NewFlagSet("edit", flag.ExitOnError)
		priority := fs.String("priority", "", "priority: low|medium|high")
		due := fs.String("due", "", "due date: YYYY-MM-DD (clear with '-')")
		doneFlagStr := fs.String("done", "", "set done: true|false")
		newName := fs.String("name", "", "new name (alternative to positional arg)")
		var tags tagList
		fs.Var(&tags, "tag", "tag (repeatable, replaces existing tags)")
		fs.Parse(cmdArgs)

		if fs.NArg() == 0 {
			fmt.Println("You must supply an ID.")
			os.Exit(1)
		}
		ids := parseIds(fs.Arg(0))
		if len(ids) == 0 {
			fmt.Println("You must supply a valid ID.")
			os.Exit(1)
		}

		item := store.GetItem(ids[0])
		if item == nil {
			fmt.Printf("Cannot find item with ID %d\n", ids[0])
			os.Exit(1)
		}

		// Determine new name.
		nameText := *newName
		if nameText == "" {
			nameText = strings.Join(fs.Args()[1:], " ")
		}
		if nameText == "" {
			var editorErr error
			nameText, editorErr = openInEditor(item.Name)
			exitOnErr(editorErr)
		}
		if nameText != "" {
			item.Name = nameText
		}

		if *priority != "" {
			item.Priority = internal.Priority(*priority)
		}
		if *due == "-" {
			item.DueDate = nil
		} else if *due != "" {
			t, err := time.Parse("2006-01-02", *due)
			if err != nil {
				fmt.Printf("Invalid date %q — use YYYY-MM-DD\n", *due)
				os.Exit(1)
			}
			item.DueDate = &t
		}
		if len(tags) > 0 {
			item.Tags = []string(tags)
		}
		if *doneFlagStr != "" {
			switch strings.ToLower(*doneFlagStr) {
			case "true", "1", "yes":
				item.Done = true
			case "false", "0", "no":
				item.Done = false
			default:
				fmt.Printf("Invalid --done value %q — use true|false\n", *doneFlagStr)
				os.Exit(1)
			}
		}

		count := store.EditItem(ids[0], *item)
		if count == 0 {
			fmt.Printf("Cannot find item with ID %d\n", ids[0])
		}

	case "done":
		if len(cmdArgs) == 0 {
			fmt.Println("You must supply an ID.")
			os.Exit(1)
		}
		ids := parseIds(cmdArgs[0])
		if len(ids) == 0 {
			fmt.Println("You must supply a valid ID.")
			os.Exit(1)
		}
		item := store.GetItem(ids[0])
		if item == nil {
			fmt.Printf("Cannot find item with ID %d\n", ids[0])
			os.Exit(1)
		}
		item.Done = true
		store.EditItem(ids[0], *item)

	case "reopen":
		if len(cmdArgs) == 0 {
			fmt.Println("You must supply an ID.")
			os.Exit(1)
		}
		ids := parseIds(cmdArgs[0])
		if len(ids) == 0 {
			fmt.Println("You must supply a valid ID.")
			os.Exit(1)
		}
		item := store.GetItem(ids[0])
		if item == nil {
			fmt.Printf("Cannot find item with ID %d\n", ids[0])
			os.Exit(1)
		}
		item.Done = false
		store.EditItem(ids[0], *item)

	case "clearall":
		store.DeleteAllItems()

	case "help", "-h", "--help":
		printHelp()

	default:
		fmt.Printf("Unknown command %s\n", cmd)
		os.Exit(1)
	}
}

func resolveMode(backendFlag string) s.Mode {
	src := backendFlag
	if src == "" {
		src = os.Getenv("TODO_BACKEND")
	}
	switch src {
	case "file":
		return s.FileMode
	case "cloud":
		return s.CloudMode
	case "sqlite", "db":
		return s.DbMode
	}
	return s.DbMode
}

func openInEditor(initial string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		return "", errors.New("set $EDITOR or $VISUAL to use editor mode")
	}

	f, err := os.CreateTemp("", "todo-*.txt")
	if err != nil {
		return "", err
	}
	f.WriteString(initial)
	f.Close()
	defer os.Remove(f.Name())

	parts := strings.Fields(editor)
	cmd := exec.Command(parts[0], append(parts[1:], f.Name())...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	content, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Todo Store")
	fmt.Println("USAGE: todo [--backend=sqlite|file|cloud] [COMMAND] [FLAGS] [ARGUMENT]")
	fmt.Println()
	fmt.Println("Commands")
	fmt.Printf("\tlist, l, ps, ls \t- list todo items\n")
	fmt.Printf("\t\t--all\t\tshow done items too\n")
	fmt.Printf("\t\t--done\t\tshow only done items\n")
	fmt.Printf("\t\t--priority\tfilter by priority: low|medium|high\n")
	fmt.Printf("\t\t--tag\t\tfilter by tag\n")
	fmt.Printf("\t\t--overdue\tshow only overdue items\n")
	fmt.Println()
	fmt.Printf("\tdelete, remove, d, rm \t- delete a todo item by id\n")
	fmt.Println()
	fmt.Printf("\tadd, create, put, a \t- add a new item\n")
	fmt.Printf("\t\t--priority\tlow|medium|high\n")
	fmt.Printf("\t\t--due\t\tYYYY-MM-DD\n")
	fmt.Printf("\t\t--tag\t\ttag (repeatable)\n")
	fmt.Printf("\t\t(no text = opens $EDITOR)\n")
	fmt.Println()
	fmt.Printf("\te, edit, update \t- edit an existing todo by id\n")
	fmt.Printf("\t\t--name\t\tnew name\n")
	fmt.Printf("\t\t--priority\tlow|medium|high\n")
	fmt.Printf("\t\t--due\t\tYYYY-MM-DD (use '-' to clear)\n")
	fmt.Printf("\t\t--tag\t\ttag (repeatable, replaces all tags)\n")
	fmt.Printf("\t\t--done\t\ttrue|false\n")
	fmt.Printf("\t\t(no text = opens $EDITOR)\n")
	fmt.Println()
	fmt.Printf("\tdone <id>\t\t- mark item as complete\n")
	fmt.Printf("\treopen <id>\t\t- mark item as not complete\n")
	fmt.Printf("\tclearall\t\t- delete all todo items\n")
	fmt.Printf("\thelp, -h, --help\t- show this help text\n")
	fmt.Println()
	fmt.Println("Environment")
	fmt.Printf("\tTODO_BACKEND=sqlite|file|cloud\t- select backend at runtime\n")
}

const (
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

func printItems(store s.TodoStore, opts s.ListOptions) {
	items := store.GetAllItems(opts)

	const maxChars = 100

	if items.MaxLengthItem >= maxChars {
		items.MaxLengthItem = maxChars + 3
	}

	nameColWidth := items.MaxLengthItem
	if nameColWidth < 4 {
		nameColWidth = 4
	}

	headerPadding := nameColWidth - 4
	if headerPadding < 0 {
		headerPadding = 0
	}

	fmt.Printf("ID\tSt   Pri  Item%s\tTags\tDue\tCreated\n",
		strings.Repeat(" ", headerPadding))

	now := time.Now()

	for _, item := range items.Items {
		doneMarker := "[ ]"
		if item.Done {
			doneMarker = "[x]"
		}

		priMarker := "[ ]"
		switch item.Priority {
		case internal.PriorityHigh:
			priMarker = "[H]"
		case internal.PriorityMedium:
			priMarker = "[M]"
		case internal.PriorityLow:
			priMarker = "[L]"
		}

		// Name column with truncation.
		nameRunes := []rune(item.Name)
		printName := string(nameRunes)
		if len(nameRunes) > maxChars {
			printName = string(nameRunes[:maxChars]) + "..."
		}
		nameLen := utf8.RuneCountInString(printName)
		namePadding := strings.Repeat(" ", nameColWidth-nameLen)

		// Tags.
		tagStr := ""
		for _, tg := range item.Tags {
			tagStr += "#" + tg + " "
		}
		tagStr = strings.TrimSpace(tagStr)

		// Due date.
		dueStr := ""
		if item.DueDate != nil {
			formatted := item.DueDate.Format("Mon 02 Jan 06")
			if item.DueDate.Before(now) {
				dueStr = ansiRed + formatted + ansiReset
			} else {
				dueStr = formatted
			}
		}

		fmt.Printf("[%d]\t%s %s  %s%s\t%s\t%s\t%s\n",
			item.ID,
			doneMarker,
			priMarker,
			printName,
			namePadding,
			tagStr,
			dueStr,
			item.CreatedAt.Format("Mon 02 Jan 06"),
		)
	}
}

func printItem(item *internal.Todo) {
	doneStr := "no"
	if item.Done {
		doneStr = "yes"
	}
	fmt.Printf("ID:\t\t%d\n", item.ID)
	fmt.Printf("Item:\t\t%s\n", item.Name)
	fmt.Printf("Done:\t\t%s\n", doneStr)
	fmt.Printf("Priority:\t%s\n", item.Priority)
	if item.DueDate != nil {
		fmt.Printf("Due:\t\t%s\n", item.DueDate.Format("Mon 02 Jan 06"))
	}
	if len(item.Tags) > 0 {
		fmt.Printf("Tags:\t\t%s\n", strings.Join(item.Tags, ", "))
	}
	fmt.Printf("Created At:\t%s\n", item.CreatedAt.Format("Mon 02 Jan 06 15:04"))
	fmt.Printf("Updated At:\t%s\n", item.UpdatedAt.Format("Mon 02 Jan 06 15:04"))
}

func parseIds(possibleIds ...string) []int {
	var ids []int

	for _, s := range possibleIds {
		id, err := strconv.Atoi(s)
		if err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}
