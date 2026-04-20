package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var pythonBinary string

var rootCmd = &cobra.Command{
	Use:   "cfs",
	Short: "CurseForge Server Installer",
	Long:  "A CLI tool for installing and updating CurseForge Minecraft server packs.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&pythonBinary, "python-binary", "",
		"Path to python3 binary (default: auto-detect)")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cfCmd)
	rootCmd.AddCommand(configCmd)
}

// resolvePython finds the Python interpreter to use.
func resolvePython() string {
	if pythonBinary != "" {
		return pythonBinary
	}
	for _, name := range []string{"python3", "python"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	fmt.Fprintln(os.Stderr, "Error: python3 not found. Install Python 3 or use --python-binary.")
	os.Exit(1)
	return ""
}

// runPython delegates to the Python mcserver CLI, forwarding stdout/stderr/stdin.
func runPython(args []string) error {
	py := resolvePython()
	fullArgs := append([]string{"-m", "mcserver.cli"}, args...)
	cmd := exec.Command(py, fullArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run from the project root so the Python package is importable.
	// If MCSERVER_PROJECT_ROOT is set, use it; otherwise use cwd.
	if root := os.Getenv("MCSERVER_PROJECT_ROOT"); root != "" {
		cmd.Dir = root
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to run python: %w", err)
	}
	return nil
}

// buildPythonArgs converts cobra flags into the equivalent Python CLI argv.
func buildPythonArgs(subcmd string, positional []string, flags map[string]string, boolFlags map[string]bool) []string {
	args := []string{subcmd}
	args = append(args, positional...)
	for k, v := range flags {
		if v != "" {
			args = append(args, "--"+k, v)
		}
	}
	for k, v := range boolFlags {
		if v {
			args = append(args, "--"+k)
		}
	}
	return args
}

// buildCfPythonArgs builds args for `cf <subcmd>` style commands.
func buildCfPythonArgs(subcmd string, positional []string, flags map[string]string, boolFlags map[string]bool) []string {
	args := []string{"cf", subcmd}
	args = append(args, positional...)
	for k, v := range flags {
		if v != "" {
			args = append(args, "--"+k, v)
		}
	}
	for k, v := range boolFlags {
		if v {
			args = append(args, "--"+k)
		}
	}
	return args
}

// buildConfigPythonArgs builds args for `config <subcmd>` style commands.
func buildConfigPythonArgs(subcmd string, positional []string) []string {
	args := []string{"config", subcmd}
	args = append(args, positional...)
	return args
}

// intToStr converts an int to string, returning "" for zero values.
func intToStr(v int) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%d", v)
}
