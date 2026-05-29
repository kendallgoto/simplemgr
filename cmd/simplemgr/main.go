package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	goutil "github.com/kendallgoto/goutil"
	"github.com/kendallgoto/simplemgr"
	"github.com/kendallgoto/simplemgr/pkg/smp"
	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

var (
	portPath string
	baudRate int
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "simplemgr",
		Short: "Manage MCUmgr devices over a serial port",
	}
	root.PersistentFlags().StringVar(&portPath, "port", "", "serial port device path (e.g. /dev/ttyACM0)")
	root.PersistentFlags().IntVar(&baudRate, "baud", 115200, "serial port baud rate")
	_ = root.MarkPersistentFlagRequired("port")

	root.AddCommand(
		newOSCmd(),
		newImageCmd(),
		newStatCmd(),
		newSettingsCmd(),
		newFSCmd(),
		newShellCmd(),
		newZephyrCmd(),
	)
	return root
}

func withPort(ctx context.Context, fn func(ctx context.Context, p *simplemgr.Port) error) error {
	dev, err := serial.Open(portPath, &serial.Mode{BaudRate: baudRate})
	if err != nil {
		return fmt.Errorf("opening serial port %q: %w", portPath, err)
	}
	defer dev.Close()
	return fn(ctx, simplemgr.New(dev))
}

func render[T any](resp T, err error) error {
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func newOSCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "os", Short: "OS management group"}

	cmd.AddCommand(&cobra.Command{
		Use:   "echo <data>",
		Short: "Echo a string back from the device",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.Echo(ctx, args[0]))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "taskstats",
		Short: "Read per-task statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetTaskStats(ctx))
			})
		},
	})

	var setDatetime string
	datetime := &cobra.Command{
		Use:   "datetime",
		Short: "Get (or --set) the device date-time value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				if setDatetime != "" {
					return render(p.SetDatetime(ctx, setDatetime))
				}
				return render(p.GetDatetime(ctx))
			})
		},
	}
	datetime.Flags().StringVar(&setDatetime, "set", "", "date to set (2026-05-28T12:00:00)") // TODO: accept wider range of formats here and convert
	cmd.AddCommand(datetime)

	var force bool
	reset := &cobra.Command{
		Use:   "reset",
		Short: "Reset the device",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				var f uint8
				if force {
					f = 1
				}
				return render(p.Reset(ctx, f))
			})
		},
	}
	reset.Flags().BoolVar(&force, "force", false, "reset even if the device reports busy")
	cmd.AddCommand(reset)

	cmd.AddCommand(&cobra.Command{
		Use:   "params",
		Short: "Read MCUmgr buffer parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetMcuMgrParameters(ctx))
			})
		},
	})

	var format string
	appinfo := &cobra.Command{
		Use:   "appinfo",
		Short: "Read OS/application info",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetOSAppInfo(ctx, format))
			})
		},
	}
	appinfo.Flags().StringVar(&format, "format", "a", "fields to return (s/n/r/v/b/m/p/i/o or a for all)") // TODO: expand field names to flags?
	cmd.AddCommand(appinfo)

	cmd.AddCommand(&cobra.Command{
		Use:   "bootloader",
		Short: "Read bootloader info",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetBootloaderInfo(ctx))
			})
		},
	})

	return cmd
}

// ============================================================================
// image
// ============================================================================

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "image", Short: "Image management group"}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List the state of images on the device",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetImageState(ctx))
			})
		},
	})

	var (
		hashHex string
		confirm bool
	)
	set := &cobra.Command{
		Use:   "set",
		Short: "Mark an image for test or confirm it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var hash []byte
			if hashHex != "" {
				h, err := hex.DecodeString(hashHex)
				if err != nil {
					return fmt.Errorf("decoding --hash: %w", err)
				}
				hash = h
			}
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.SetImageState(ctx, hash, confirm))
			})
		},
	}
	set.Flags().StringVar(&hashHex, "hash", "", "hex-encoded image hash (defaults to the running image)")
	set.Flags().BoolVar(&confirm, "confirm", false, "confirm the image (otherwise mark for test)")
	cmd.AddCommand(set)

	var (
		imageNum  int
		chunkSize int
	)
	upload := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload (flash) a firmware image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("reading image: %w", err)
			}
			sum := sha256.Sum256(data)
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				off := uint32(0)
				total := uint32(len(data))
				for off < total {
					end := off + uint32(chunkSize)
					if end > total {
						end = total
					}
					req := &smp.ImageUploadRequest{
						Offset: off,
						Data:   data[off:end],
					}
					if off == 0 {
						req.Length = goutil.Ptr(total)
						req.Hash = sum[:]
						if imageNum >= 0 {
							req.Image = goutil.Ptr(uint32(imageNum))
						}
					}
					resp, err := p.UploadImage(ctx, req)
					if err != nil {
						return fmt.Errorf("upload at offset %d: %w", off, err)
					}
					if resp.Offset <= off {
						return fmt.Errorf("device did not advance offset (stuck at %d)", off)
					}
					off = resp.Offset
					fmt.Fprintf(os.Stderr, "\ruploaded %d/%d bytes", off, total)
				}
				fmt.Fprintln(os.Stderr)
				return nil
			})
		},
	}
	upload.Flags().IntVar(&imageNum, "image", 0, "target image number")
	upload.Flags().IntVar(&chunkSize, "chunk", 128, "upload chunk size in bytes")
	cmd.AddCommand(upload)

	var slot int
	erase := &cobra.Command{
		Use:   "erase",
		Short: "Erase an image slot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.EraseImage(ctx, uint8(slot)))
			})
		},
	}
	erase.Flags().IntVar(&slot, "slot", 1, "slot to erase")
	cmd.AddCommand(erase)

	return cmd
}

// ============================================================================
// stat
// ============================================================================

func newStatCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "stat", Short: "Statistics management group"}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available statistics groups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetStatListGroups(ctx))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <group>",
		Short: "Read the data for a statistics group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetStatGroupData(ctx, args[0]))
			})
		},
	})

	return cmd
}

// ============================================================================
// settings
// ============================================================================

func newSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "settings", Short: "Settings (config) management group"}

	var maxSize uint64
	read := &cobra.Command{
		Use:   "read <name>",
		Short: "Read a setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.ReadSetting(ctx, args[0], maxSize))
			})
		},
	}
	read.Flags().Uint64Var(&maxSize, "max-size", 0, "maximum value size to return")
	cmd.AddCommand(read)

	cmd.AddCommand(&cobra.Command{
		Use:   "write <name> <value>",
		Short: "Write a setting",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.WriteSetting(ctx, args[0], []byte(args[1])))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.DeleteSetting(ctx, args[0]))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "commit",
		Short: "Commit pending settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.CommitSettings(ctx))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "load",
		Short: "Load settings from persistent storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.LoadSettings(ctx))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "save",
		Short: "Save settings to persistent storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.SaveSettings(ctx))
			})
		},
	})

	return cmd
}

// ============================================================================
// fs
// ============================================================================

func newFSCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "fs", Short: "File management group"}

	var dlChunk int
	download := &cobra.Command{
		Use:   "download <remote-path> <local-path>",
		Short: "Download a file from the device",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			remote, local := args[0], args[1]
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				var (
					buf   []byte
					off   uint64
					total uint64
				)
				for {
					resp, err := p.DownloadFile(ctx, remote, off)
					if err != nil {
						return fmt.Errorf("download at offset %d: %w", off, err)
					}
					if off == 0 && resp.Length != nil {
						total = *resp.Length
					}
					buf = append(buf, resp.Data...)
					off += uint64(len(resp.Data))
					fmt.Fprintf(os.Stderr, "\rdownloaded %d/%d bytes", off, total)
					if len(resp.Data) == 0 || off >= total {
						break
					}
				}
				fmt.Fprintln(os.Stderr)
				if err := os.WriteFile(local, buf, 0o644); err != nil {
					return fmt.Errorf("writing %q: %w", local, err)
				}
				return nil
			})
		},
	}
	download.Flags().IntVar(&dlChunk, "chunk", 128, "download chunk size hint (unused; device controls chunking)")
	cmd.AddCommand(download)

	var ulChunk int
	upload := &cobra.Command{
		Use:   "upload <local-path> <remote-path>",
		Short: "Upload a file to the device",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			local, remote := args[0], args[1]
			data, err := os.ReadFile(local)
			if err != nil {
				return fmt.Errorf("reading %q: %w", local, err)
			}
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				off := uint64(0)
				total := uint64(len(data))
				for off < total {
					end := off + uint64(ulChunk)
					if end > total {
						end = total
					}
					req := &smp.FsUploadRequest{
						Name:   remote,
						Offset: off,
						Data:   data[off:end],
					}
					if off == 0 {
						req.Length = goutil.Ptr(total)
					}
					resp, err := p.UploadFile(ctx, req)
					if err != nil {
						return fmt.Errorf("upload at offset %d: %w", off, err)
					}
					if resp.Offset <= off {
						return fmt.Errorf("device did not advance offset (stuck at %d)", off)
					}
					off = resp.Offset
					fmt.Fprintf(os.Stderr, "\ruploaded %d/%d bytes", off, total)
				}
				fmt.Fprintln(os.Stderr)
				return nil
			})
		},
	}
	upload.Flags().IntVar(&ulChunk, "chunk", 128, "upload chunk size in bytes")
	cmd.AddCommand(upload)

	cmd.AddCommand(&cobra.Command{
		Use:   "status <remote-path>",
		Short: "Get the size of a file on the device",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetFileStatus(ctx, args[0]))
			})
		},
	})

	var (
		hashType string
		hashOff  uint64
		hashLen  uint64
	)
	hash := &cobra.Command{
		Use:   "hash <remote-path>",
		Short: "Compute a hash/checksum of a file on the device",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &smp.FsHashRequest{Name: args[0]}
			if hashType != "" {
				req.Type = goutil.Ptr(hashType)
			}
			if cmd.Flags().Changed("off") {
				req.Offset = goutil.Ptr(hashOff)
			}
			if cmd.Flags().Changed("len") {
				req.Length = goutil.Ptr(hashLen)
			}
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetFileHash(ctx, req))
			})
		},
	}
	hash.Flags().StringVar(&hashType, "type", "", "hash/checksum type (e.g. crc32, sha256)")
	hash.Flags().Uint64Var(&hashOff, "off", 0, "offset to start hashing at")
	hash.Flags().Uint64Var(&hashLen, "len", 0, "number of bytes to hash")
	cmd.AddCommand(hash)

	cmd.AddCommand(&cobra.Command{
		Use:   "hashtypes",
		Short: "List supported hash/checksum types",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.GetSupportedFileHashTypes(ctx))
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "close",
		Short: "Close any open file handle on the device",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.CloseFile(ctx))
			})
		},
	})

	return cmd
}

// ============================================================================
// shell
// ============================================================================

func newShellCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "shell", Short: "Shell management group"}

	cmd.AddCommand(&cobra.Command{
		Use:   "exec <arg>...",
		Short: "Execute a shell command line on the device",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.Execute(ctx, args))
			})
		},
	})

	return cmd
}

// ============================================================================
// zephyr
// ============================================================================

func newZephyrCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "zephyr", Short: "Zephyr-specific management group"}

	cmd.AddCommand(&cobra.Command{
		Use:   "erase",
		Short: "Erase the storage partition",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withPort(cmd.Context(), func(ctx context.Context, p *simplemgr.Port) error {
				return render(p.EraseStorage(ctx))
			})
		},
	})

	return cmd
}
