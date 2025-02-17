// Copyright © 2016-2021 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"

	"github.com/dustin/go-humanize"
	"github.com/tatsushid/go-prettytable"
)

// dimCmd represents the stat command
var dimCmd = &cobra.Command{
	Use:     "dim",
	Aliases: []string{"size", "stats", "stat"},
	Short:   "dimensions of CSV file",
	Long: `dimensions of CSV file

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		tabular := getFlagBool(cmd, "tabular")
		cols := getFlagBool(cmd, "cols")
		rows := getFlagBool(cmd, "rows")
		noFiles := getFlagBool(cmd, "no-files")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		var tbl *prettytable.Table
		if rows || cols {
		} else if tabular {
			outfh.WriteString("file\tnum_cols\tnum_rows\n")
		} else {
			tbl, err = prettytable.NewTable([]prettytable.Column{
				{Header: "file"},
				{Header: "num_cols", AlignRight: true},
				{Header: "num_rows", AlignRight: true}}...)
			checkError(err)
			tbl.Separator = "   "
		}

		for _, file := range files {
			var numCols, numRows uint64

			csvReader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					if rows {
						if noFiles {
							outfh.WriteString(fmt.Sprintf("%d\n", numRows))
						} else {
							outfh.WriteString(fmt.Sprintf("%s\t%d\n", file, numRows))
						}
					} else if tabular {
						if noFiles {
							outfh.WriteString(fmt.Sprintf("%d\t%d\n", numCols, numRows))
						} else {
							outfh.WriteString(fmt.Sprintf("%s\t%d\t%d\n", file, numCols, numRows))
						}
					} else {
						tbl.AddRow(
							file,
							humanize.Comma(int64(numCols)),
							humanize.Comma(int64(numRows)))
					}

					continue
				} else {
					checkError(err)
				}
			}

			csvReader.Run()

			once := true
		HERE:
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				numRows += uint64(len(chunk.Data))
				if once {
					for _, record := range chunk.Data {
						numCols = uint64(len(record))
						break
					}
					if cols {
						if noFiles {
							outfh.WriteString(fmt.Sprintf("%d\n", numCols))
						} else {
							outfh.WriteString(fmt.Sprintf("%s\t%d\n", file, numCols))
						}
						break HERE
					}
					once = false
				}
			}
			if cols {
				continue
			}
			if numRows > 0 && !config.NoHeaderRow {
				numRows--
			}

			if rows {
				if noFiles {
					outfh.WriteString(fmt.Sprintf("%d\n", numRows))
				} else {
					outfh.WriteString(fmt.Sprintf("%s\t%d\n", file, numRows))
				}
			} else if tabular {
				if noFiles {
					outfh.WriteString(fmt.Sprintf("%d\t%d\n", numCols, numRows))
				} else {
					outfh.WriteString(fmt.Sprintf("%s\t%d\t%d\n", file, numCols, numRows))
				}
			} else {
				tbl.AddRow(
					file,
					humanize.Comma(int64(numCols)),
					humanize.Comma(int64(numRows)))
			}

			readerReport(&config, csvReader, file)

		}
		if !(rows || cols) && !tabular {
			outfh.Write(tbl.Bytes())
		}
	},
}

func init() {
	dimCmd.Flags().BoolP("tabular", "", false, `output in machine-friendly tabular format`)
	dimCmd.Flags().BoolP("cols", "", false, `only print number of columns (or using "csvtk ncol"`)
	dimCmd.Flags().BoolP("rows", "", false, `only print number of rows (or using "csvtk nrow")`)
	dimCmd.Flags().BoolP("no-files", "n", false, "do not print file names (only affect --cols and --rows)")

	RootCmd.AddCommand(dimCmd)

}
