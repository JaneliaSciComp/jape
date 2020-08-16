package cmd

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"io"
	Utils "jape/utils"
	"os"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the container for the current project",
	Long: `Runs a docker container for the current Jape project. The current directory must contain a jape.yaml 
file describing the project. You can create a runnable project using the init and build commands.
`,
	Run: func(cmd *cobra.Command, args []string) {
		run(args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	// Disable parsing because we want to pass through flags to the containerized application
	runCmd.DisableFlagParsing = true
}

func run(args []string) {

	config := Utils.ReadProjectConfig()
	versionTag := config.Name+":"+config.Version

	Utils.PrintInfo("Running %s", versionTag)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		Utils.PrintFatal("%s", err)
	}
	defer cli.Close()

	ctx := context.Background()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        versionTag,
		Cmd:          args,
		Tty:          true,
		AttachStdin:  true,
	}, nil, nil, nil, "")
	if err != nil {
		Utils.PrintFatal("%s", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		Utils.PrintFatal("%s", err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			Utils.PrintFatal("%s", err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
	})
	if err != nil {
		Utils.PrintFatal("%s", err)
	}

	if _, err := io.Copy(os.Stdout, out); err != nil {
		Utils.PrintFatal("%s", err)
	}
}