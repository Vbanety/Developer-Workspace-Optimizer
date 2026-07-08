// Command devopt is the Developer Workspace Optimizer CLI: scans dev-tool
// caches, reports recoverable space, and cleans safely with confirmation.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vinidev/devopt/internal/config"
	"github.com/vinidev/devopt/internal/core"
	"github.com/vinidev/devopt/internal/modules/linux"
	"github.com/vinidev/devopt/internal/report"
	"github.com/vinidev/devopt/internal/tui"
)

// version is set at release build time; "dev" during local builds.
var version = "0.1.0-dev"

// linuxModules lists every v0.1 module: name and whether it's safe to clean
// without extra confirmation (see .claude/contexts/safety.md). apt lives at
// a system path (/var/cache/apt/archives) and needs root to remove — never
// mark it safe, even though mechanically it's the same DirCache shape.
var linuxModules = []struct {
	name string
	safe bool
}{
	{"yarn", true},
	{"npm", true},
	{"pnpm", true},
	{"gradle", true},
	{"composer", true},
	{"playwright", true},
	{"puppeteer", true},
	{"trash", true},
	{"apt", false},
}

func main() {
	root := &cobra.Command{
		Use:   "devopt",
		Short: "Developer Workspace Optimizer — escaneia e limpa caches de dev com segurança",
		// Bare invocation (no subcommand) launches the interactive menu.
		// `report`/`clean` remain the scriptable, non-interactive path.
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := buildRegistry()
			if err != nil {
				return err
			}
			if err := tui.Run(reg); err != nil {
				return fmt.Errorf(
					"menu interativo indisponível (%w) — sem terminal interativo? use 'devopt report' ou 'devopt clean'",
					err,
				)
			}
			return nil
		},
		SilenceUsage: true, // runtime errors here aren't a CLI-usage mistake — don't dump the flag/command list over them
	}
	root.AddCommand(newReportCmd(), newCleanCmd(), newVersionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Mostra a versão do devopt",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("devopt " + version)
			return nil
		},
	}
}

// buildRegistry wires up every module for the current OS. v0.1 only
// supports Linux — see .claude/contexts/roadmap.md for Windows/macOS phases.
func buildRegistry() (*core.Registry, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("devopt v0.1 só suporta Linux (detectado: %s)", runtime.GOOS)
	}

	rules, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("carregando config: %w", err)
	}

	reg := core.NewRegistry()
	for _, m := range linuxModules {
		reg.Register(linux.NewDirCache(m.name, rules.Paths[m.name], m.safe))
	}

	// Bespoke modules that aren't a plain cache directory (see
	// .claude/contexts/architecture.md for why each needs its own type).
	reg.Register(linux.NewDocker())
	reg.Register(linux.NewSnap())
	reg.Register(linux.NewMultiDirCache("cursor", rules.Paths["cursor"], linux.ElectronCacheSubdirs, true))
	reg.Register(linux.NewMultiDirCache("vscode", rules.Paths["vscode"], linux.ElectronCacheSubdirs, true))

	return reg, nil
}

func newReportCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Escaneia caches e mostra espaço recuperável (nunca apaga nada)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := buildRegistry()
			if err != nil {
				return err
			}
			rows := report.Scan(reg)
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}
			cmd.Print(report.Render(rows))
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "saída em JSON")
	return cmd
}

func newCleanCmd() *cobra.Command {
	var (
		safeOnly  bool
		deep      bool
		moduleArg string
		dryRun    bool
		yes       bool
	)
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Limpa caches detectados, com relatório, confirmação e guard de segurança antes de qualquer remoção",
		RunE: func(cmd *cobra.Command, _ []string) error {
			reg, err := buildRegistry()
			if err != nil {
				return err
			}

			targets := reg.All()
			if moduleArg != "" {
				m := reg.Get(moduleArg)
				if m == nil {
					return fmt.Errorf("módulo desconhecido: %s", moduleArg)
				}
				targets = []core.Module{m}
			}
			// --deep doesn't change targeting today (no --safe already means
			// "every detected module"); kept as a distinct flag so a future,
			// more aggressive mode (e.g. `docker system prune -a`) has
			// somewhere to hang without breaking scripts written against it.
			_ = deep

			for _, m := range targets {
				if err := cleanOne(cmd, m, safeOnly, dryRun, yes); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&safeOnly, "safe", false, "só módulos marcados como seguros")
	cmd.Flags().BoolVar(&deep, "deep", false, "limpeza profunda (reservado para v0.2)")
	cmd.Flags().StringVar(&moduleArg, "module", "", "limitar a um módulo específico")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "mostra o que seria feito, não apaga nada")
	cmd.Flags().BoolVar(&yes, "yes", false, "pula confirmação por módulo")
	return cmd
}

func cleanOne(cmd *cobra.Command, m core.Module, safeOnly, dryRun, yes bool) error {
	found, err := m.Detect()
	if err != nil {
		return fmt.Errorf("%s: %w", m.Name(), err)
	}
	if !found {
		return nil
	}
	if safeOnly && !m.Safe() {
		return nil
	}

	finding, err := m.Calculate()
	if err == core.ErrNotImplemented {
		cmd.Printf("… %s: não implementado, pulando\n", m.Name())
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", m.Name(), err)
	}

	cmd.Printf("%s: %s recuperável em %s\n", m.Name(), report.HumanSize(finding.SizeBytes), finding.Path)

	if !yes && !confirm(cmd, m.Name()) {
		cmd.Printf("… %s: pulado (não confirmado)\n", m.Name())
		return nil
	}

	result, err := m.Clean(dryRun)
	if err != nil {
		return fmt.Errorf("%s: %w", m.Name(), err)
	}
	if result.Skipped {
		cmd.Printf("… %s: pulado (%s)\n", m.Name(), result.SkipReason)
		return nil
	}

	verb := "liberado"
	if dryRun {
		verb = "seria liberado (dry-run)"
	}
	cmd.Printf("✔ %s: %s %s\n", m.Name(), report.HumanSize(result.FreedBytes), verb)
	return nil
}

func confirm(cmd *cobra.Command, name string) bool {
	cmd.Printf("Deseja remover %s? (S/n) ", name)
	reader := bufio.NewReader(cmd.InOrStdin())
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "" || line == "s" || line == "y" || line == "sim" || line == "yes"
}
