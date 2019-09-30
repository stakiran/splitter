package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ProductName = "splitter"
const ProductVersion = "0.0.1"

func ____util___() {
}

func success() {
	os.Exit(0)
}

func abort(msg string) {
	fmt.Printf("[Error!] %s\n", msg)
	os.Exit(1)
}

func warn(msg string) {
	fmt.Printf("[Warning!] %s\n", msg)
}

func debugprint(useDebugprint bool, msg string) {
	if useDebugprint {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

func file2list(filepath string) []string {
	fp, err := os.Open(filepath)
	if err != nil {
		abort(err.Error())
	}
	defer fp.Close()

	lines := []string{}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines
}

func list2file(filepath string, lines []string) {
	fp, err := os.Create(filepath)
	if err != nil {
		abort(err.Error())
	}
	defer fp.Close()

	writer := bufio.NewWriter(fp)
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
}

func list2fileRelatively(dirname string, filename string, lines []string){
	fullpath := filepath.Join(dirname, filename)
	list2file(fullpath, lines)
}

func isExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func isExistingDirectory(filepath string) bool {
	if !isExist(filepath) {
		return false
	}
	info, _ := os.Stat(filepath)
	return info.IsDir()
}

func ____classes____() {
}

type FileInfo struct {
	LineCount int
}

type IndexSaver struct {
	filenames    []string
	fileinfos    []FileInfo
	selfFilename string
	outputDir string
}

func NewIndexSaver(filename string, outputDir string) IndexSaver {
	inst := IndexSaver{}
	inst.selfFilename = filename
	inst.outputDir = outputDir
	return inst
}

func (saver *IndexSaver) onRefresh(filename string, fileinfo FileInfo){
	saver.addFileInformation(filename, fileinfo)
}

func (saver *IndexSaver) addFileInformation(filename string, fileinfo FileInfo) {
	saver.filenames = append(saver.filenames, filename)
	saver.fileinfos = append(saver.fileinfos, fileinfo)
}

func (saver *IndexSaver) Save() {
	contents := []string{}

	sectionInterval := 30
	lineCountGraphMark := "@"
	lineCountGraphUnit := 30
	lineCountGraphLimitLength := 20
	blankLine := ""

	for i, filename := range saver.filenames {
		if i%sectionInterval==0{
			sectionLine := fmt.Sprintf("# %s", filename)
			contents = append(contents, blankLine)
			contents = append(contents, sectionLine)
			contents = append(contents, blankLine)

			tableHeader1 := "| Lines | Title |"
			tableHeader2 := "| ----- | ----- |"
			contents = append(contents, tableHeader1, tableHeader2)
		}

		curFileInfo := saver.fileinfos[i]

		lineCount := curFileInfo.LineCount
		lineCountGraph := strings.Repeat(lineCountGraphMark, lineCount/lineCountGraphUnit)
		if len(lineCountGraph)>lineCountGraphLimitLength{
			lineCountGraph = strings.Repeat(lineCountGraphMark, lineCountGraphLimitLength)
		}
		if len(lineCountGraph)==0{
			lineCountGraph = lineCountGraphMark
		}

		content := fmt.Sprintf("| %s | [%s](%s) |", lineCountGraph, filename, filename)
		contents = append(contents, content)
	}

	list2fileRelatively(saver.outputDir, saver.selfFilename, contents)
}

type RefreshCallbackee interface{
	onRefresh(string, FileInfo)
}

type LinesSaver struct {
	lines    []string
	filename string
	outputDir string
	onRefreshCallbackees   []RefreshCallbackee
}

func NewLinesSaver(outputDir string) LinesSaver {
	inst := LinesSaver{}
	inst.outputDir = outputDir
	return inst
}

func (saver *LinesSaver) AddCallbackOnRefresh(obj RefreshCallbackee){
	saver.onRefreshCallbackees = append(saver.onRefreshCallbackees, obj)
}

func (saver *LinesSaver) Refresh(newFilename string) {
	// ここまで溜まっているであろう内容を処理した後, 新しいファイル名をセット.
	if saver.filename != "" {
		list2fileRelatively(saver.outputDir, saver.filename, saver.lines)

		fileinfo := FileInfo{}
		fileinfo.LineCount = len(saver.lines)
		for _, obj := range saver.onRefreshCallbackees {
			obj.onRefresh(saver.filename, fileinfo)
		}
	}

	saver.filename = newFilename
	saver.lines = []string{}
}

func (saver *LinesSaver) AppendLine(line string) {
	saver.lines = append(saver.lines, line)
}

func ____funcs____() {
}

func isLevel1Line(line string) bool {
	peekFirst := string([]rune(line)[:1])
	peekSecond := string([]rune(line)[1:2])
	sectionChar := "#"
	return peekFirst == sectionChar && peekSecond != sectionChar
}

func isHilightLine(line string) bool {
	peek := string([]rune(line)[:3])
	hilightGrammer := "```"
	return peek == hilightGrammer
}

func sectionname2filename(sectionName string) string {
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	AfterChar := "-"
	ReplaceCountIsUnlimited := -1
	Extension := ".md"

	ret := ""

	ret = strings.TrimSpace(sectionName)

	// replace invalid chars on windows filename
	for _, invalidChar := range invalidChars {
		ret = strings.Replace(ret, invalidChar, AfterChar, ReplaceCountIsUnlimited)
	}

	// replace spaces for easy to handling as a filename
	ret = strings.Replace(ret, " ", AfterChar, ReplaceCountIsUnlimited)
	ret = strings.Replace(ret, "\t", AfterChar, ReplaceCountIsUnlimited)

	// replace () because the conflict on markdown grammer
	ret = strings.Replace(ret, "(", AfterChar, ReplaceCountIsUnlimited)
	ret = strings.Replace(ret, ")", AfterChar, ReplaceCountIsUnlimited)

	// add the extension
	ret = ret + Extension

	return ret
}

func line2sectionname(line string) string {
	return string([]rune(line)[1:])
}

func ____arguments____(){
}

type Args struct {
	outputDir *string
	targetFilepath *string
}

func argparse() Args {
	args := Args{}

	args.targetFilepath = flag.String("targetFilepath", "", "A target markdown file path to split.")
	args.outputDir = flag.String("outputDirectory", "docs", "An output directory(must already exists) based on the current directory relative. ")
	isShowingVersion := flag.Bool("version", false, "Show this tool version.")

	flag.Parse()

	if *isShowingVersion {
		fmt.Printf("%s %s\n", ProductName, ProductVersion)
		success()
	}

	return args
}

func main() {
	args := argparse()
	outputDir := *args.outputDir
	targetFilepath := *args.targetFilepath

	if targetFilepath == ""{
		abort("A targetFilepath required.")
	}

	lines := file2list(targetFilepath)

	// index saver のアルゴリズム上, 最後の見出しを捉えられないので
	// 末尾にダミーの見出しを入れることで回避.
	lines = append(lines, "# dummy to avoiding loss of the last section")

	isInHilight := false
	linesSaver := NewLinesSaver(outputDir)
	indexSaver := NewIndexSaver("index.md", outputDir)

	var refreshCallbackeeIndexSaver RefreshCallbackee = &indexSaver
	linesSaver.AddCallbackOnRefresh(refreshCallbackeeIndexSaver)

	for _, line := range lines {

		if isHilightLine(line) {
			// ``` ← これが奇数個しか存在しないケースは想定してない
			if !isInHilight {
				isInHilight = true
			} else {
				isInHilight = false
			}
		}

		if isLevel1Line(line) && !isInHilight {
			filename := sectionname2filename(line2sectionname(line))
			linesSaver.Refresh(filename)
			// 大見出しの内容も含めたい
			linesSaver.AppendLine(line)
			continue
		}

		linesSaver.AppendLine(line)
	}

	indexSaver.Save()

}
