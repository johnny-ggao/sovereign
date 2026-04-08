package excelize

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

type File struct {
	sheetOrder []string
	sheets     map[string]*sheetData
}

type sheetData struct {
	cells     map[int]map[int]string
	colWidths []columnWidth
}

type columnWidth struct {
	Start int
	End   int
	Width float64
}

func NewFile() *File {
	f := &File{
		sheetOrder: []string{"Sheet1"},
		sheets:     map[string]*sheetData{},
	}
	f.sheets["Sheet1"] = newSheetData()
	return f
}

func OpenReader(r io.Reader) (*File, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read xlsx: %w", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open xlsx zip: %w", err)
	}

	files := make(map[string]*zip.File, len(reader.File))
	for _, file := range reader.File {
		files[file.Name] = file
	}

	workbookBytes, err := readZipFile(files, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	relsBytes, err := readZipFile(files, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}
	sharedStrings, err := readSharedStrings(files)
	if err != nil {
		return nil, err
	}

	var workbook workbookXML
	if err := xml.Unmarshal(workbookBytes, &workbook); err != nil {
		return nil, fmt.Errorf("parse workbook: %w", err)
	}
	var rels relationshipsXML
	if err := xml.Unmarshal(relsBytes, &rels); err != nil {
		return nil, fmt.Errorf("parse workbook rels: %w", err)
	}

	targets := make(map[string]string, len(rels.Relationships))
	for _, rel := range rels.Relationships {
		targets[rel.ID] = path.Clean(path.Join("xl", rel.Target))
	}

	file := &File{
		sheetOrder: make([]string, 0, len(workbook.Sheets)),
		sheets:     make(map[string]*sheetData, len(workbook.Sheets)),
	}
	for _, sheet := range workbook.Sheets {
		target, ok := targets[sheet.RelID]
		if !ok {
			return nil, fmt.Errorf("missing sheet relation for %s", sheet.Name)
		}
		sheetBytes, err := readZipFile(files, target)
		if err != nil {
			return nil, err
		}
		parsedSheet, err := parseSheet(sheetBytes, sharedStrings)
		if err != nil {
			return nil, err
		}
		file.sheetOrder = append(file.sheetOrder, sheet.Name)
		file.sheets[sheet.Name] = parsedSheet
	}
	if len(file.sheetOrder) == 0 {
		return nil, fmt.Errorf("xlsx contains no worksheets")
	}
	return file, nil
}

func (f *File) GetSheetList() []string {
	return append([]string(nil), f.sheetOrder...)
}

func (f *File) SetSheetName(oldName, newName string) {
	if oldName == newName || newName == "" {
		return
	}
	sheet, ok := f.sheets[oldName]
	if !ok {
		return
	}
	delete(f.sheets, oldName)
	f.sheets[newName] = sheet
	for i, name := range f.sheetOrder {
		if name == oldName {
			f.sheetOrder[i] = newName
		}
	}
}

func (f *File) SetCellValue(sheet, axis string, value interface{}) error {
	sheetData, ok := f.sheets[sheet]
	if !ok {
		return fmt.Errorf("sheet %s not found", sheet)
	}
	row, col, err := splitCellRef(axis)
	if err != nil {
		return err
	}
	if sheetData.cells[row] == nil {
		sheetData.cells[row] = map[int]string{}
	}
	sheetData.cells[row][col] = fmt.Sprint(value)
	return nil
}

func (f *File) SetColWidth(sheet, startCol, endCol string, width float64) error {
	sheetData, ok := f.sheets[sheet]
	if !ok {
		return fmt.Errorf("sheet %s not found", sheet)
	}
	start, err := colLettersToIndex(startCol)
	if err != nil {
		return err
	}
	end, err := colLettersToIndex(endCol)
	if err != nil {
		return err
	}
	sheetData.colWidths = append(sheetData.colWidths, columnWidth{Start: start, End: end, Width: width})
	return nil
}

func (f *File) GetRows(sheet string) ([][]string, error) {
	sheetData, ok := f.sheets[sheet]
	if !ok {
		return nil, fmt.Errorf("sheet %s not found", sheet)
	}
	maxRow := 0
	maxCol := 0
	for row, cols := range sheetData.cells {
		if row > maxRow {
			maxRow = row
		}
		for col := range cols {
			if col > maxCol {
				maxCol = col
			}
		}
	}
	rows := make([][]string, 0, maxRow)
	for rowIdx := 1; rowIdx <= maxRow; rowIdx++ {
		row := make([]string, maxCol)
		for colIdx, value := range sheetData.cells[rowIdx] {
			row[colIdx-1] = value
		}
		rows = append(rows, trimTrailingEmpty(row))
	}
	return rows, nil
}

func (f *File) Write(w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	for _, entry := range buildWorkbookEntries(f) {
		writer, err := zipWriter.Create(entry.Name)
		if err != nil {
			return fmt.Errorf("create xlsx entry %s: %w", entry.Name, err)
		}
		if _, err := writer.Write([]byte(entry.Body)); err != nil {
			return fmt.Errorf("write xlsx entry %s: %w", entry.Name, err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("finalize xlsx: %w", err)
	}
	return nil
}

type workbookXML struct {
	Sheets []workbookSheetXML `xml:"sheets>sheet"`
}

type workbookSheetXML struct {
	Name  string `xml:"name,attr"`
	RelID string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type relationshipsXML struct {
	Relationships []relationshipXML `xml:"Relationship"`
}

type relationshipXML struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type worksheetXML struct {
	Rows []worksheetRowXML `xml:"sheetData>row"`
}

type worksheetRowXML struct {
	Index int                `xml:"r,attr"`
	Cells []worksheetCellXML `xml:"c"`
}

type worksheetCellXML struct {
	Ref       string             `xml:"r,attr"`
	Type      string             `xml:"t,attr"`
	Value     string             `xml:"v"`
	InlineStr worksheetInlineXML `xml:"is"`
}

type worksheetInlineXML struct {
	Text string `xml:"t"`
}

type sharedStringsXML struct {
	Items []sharedStringItemXML `xml:"si"`
}

type sharedStringItemXML struct {
	Text string                `xml:"t"`
	Runs []sharedStringRunXML  `xml:"r"`
}

type sharedStringRunXML struct {
	Text string `xml:"t"`
}

type workbookEntry struct {
	Name string
	Body string
}

func newSheetData() *sheetData {
	return &sheetData{cells: map[int]map[int]string{}}
}

func readZipFile(files map[string]*zip.File, name string) ([]byte, error) {
	file, ok := files[name]
	if !ok {
		return nil, fmt.Errorf("xlsx entry %s not found", name)
	}
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open xlsx entry %s: %w", name, err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read xlsx entry %s: %w", name, err)
	}
	return data, nil
}

func readSharedStrings(files map[string]*zip.File) ([]string, error) {
	file, ok := files["xl/sharedStrings.xml"]
	if !ok {
		return nil, nil
	}
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open shared strings: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read shared strings: %w", err)
	}

	var shared sharedStringsXML
	if err := xml.Unmarshal(data, &shared); err != nil {
		return nil, fmt.Errorf("parse shared strings: %w", err)
	}
	values := make([]string, 0, len(shared.Items))
	for _, item := range shared.Items {
		if item.Text != "" {
			values = append(values, item.Text)
			continue
		}
		var builder strings.Builder
		for _, run := range item.Runs {
			builder.WriteString(run.Text)
		}
		values = append(values, builder.String())
	}
	return values, nil
}

func parseSheet(data []byte, sharedStrings []string) (*sheetData, error) {
	var worksheet worksheetXML
	if err := xml.Unmarshal(data, &worksheet); err != nil {
		return nil, fmt.Errorf("parse worksheet: %w", err)
	}
	sheet := newSheetData()
	for _, row := range worksheet.Rows {
		if sheet.cells[row.Index] == nil {
			sheet.cells[row.Index] = map[int]string{}
		}
		for _, cell := range row.Cells {
			_, col, err := splitCellRef(cell.Ref)
			if err != nil {
				return nil, err
			}
			sheet.cells[row.Index][col] = resolveCellValue(cell, sharedStrings)
		}
	}
	return sheet, nil
}

func resolveCellValue(cell worksheetCellXML, sharedStrings []string) string {
	switch cell.Type {
	case "s":
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return ""
		}
		return sharedStrings[index]
	case "inlineStr":
		return cell.InlineStr.Text
	default:
		return cell.Value
	}
}

func buildWorkbookEntries(f *File) []workbookEntry {
	entries := []workbookEntry{
		{Name: "[Content_Types].xml", Body: buildContentTypesXML(len(f.sheetOrder))},
		{Name: "_rels/.rels", Body: relsXML},
		{Name: "xl/workbook.xml", Body: buildWorkbookXML(f.sheetOrder)},
		{Name: "xl/_rels/workbook.xml.rels", Body: buildWorkbookRelsXML(len(f.sheetOrder))},
		{Name: "xl/styles.xml", Body: stylesXML},
	}
	for idx, sheetName := range f.sheetOrder {
		entries = append(entries, workbookEntry{
			Name: fmt.Sprintf("xl/worksheets/sheet%d.xml", idx+1),
			Body: buildSheetXML(f.sheets[sheetName]),
		})
	}
	return entries
}

func buildContentTypesXML(sheetCount int) string {
	var overrides strings.Builder
	for i := 1; i <= sheetCount; i++ {
		overrides.WriteString(fmt.Sprintf(
			`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`,
			i,
		))
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		overrides.String() +
		`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
		`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` +
		`</Types>`
}

func buildWorkbookXML(sheetNames []string) string {
	var sheets strings.Builder
	for idx, name := range sheetNames {
		sheets.WriteString(fmt.Sprintf(
			`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`,
			xmlEscape(name),
			idx+1,
			idx+1,
		))
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<sheets>` + sheets.String() + `</sheets></workbook>`
}

func buildWorkbookRelsXML(sheetCount int) string {
	var rels strings.Builder
	for i := 1; i <= sheetCount; i++ {
		rels.WriteString(fmt.Sprintf(
			`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`,
			i,
			i,
		))
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		rels.String() + `</Relationships>`
}

func buildSheetXML(sheet *sheetData) string {
	rows := make([]int, 0, len(sheet.cells))
	for row := range sheet.cells {
		rows = append(rows, row)
	}
	sort.Ints(rows)

	var colsXML strings.Builder
	if len(sheet.colWidths) > 0 {
		colsXML.WriteString("<cols>")
		for _, colWidth := range sheet.colWidths {
			colsXML.WriteString(fmt.Sprintf(
				`<col min="%d" max="%d" width="%g" customWidth="1"/>`,
				colWidth.Start,
				colWidth.End,
				colWidth.Width,
			))
		}
		colsXML.WriteString("</cols>")
	}

	var rowsXML strings.Builder
	for _, row := range rows {
		cols := make([]int, 0, len(sheet.cells[row]))
		for col := range sheet.cells[row] {
			cols = append(cols, col)
		}
		sort.Ints(cols)
		rowsXML.WriteString(fmt.Sprintf(`<row r="%d">`, row))
		for _, col := range cols {
			ref := fmt.Sprintf("%s%d", indexToColLetters(col), row)
			rowsXML.WriteString(
				fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, ref, xmlEscape(sheet.cells[row][col])),
			)
		}
		rowsXML.WriteString(`</row>`)
	}

	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
		colsXML.String() +
		`<sheetData>` + rowsXML.String() + `</sheetData></worksheet>`
}

func splitCellRef(ref string) (int, int, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return 0, 0, fmt.Errorf("empty cell reference")
	}
	letters := strings.Builder{}
	numbers := strings.Builder{}
	for _, r := range ref {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			if numbers.Len() > 0 {
				return 0, 0, fmt.Errorf("invalid cell reference %s", ref)
			}
			letters.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			numbers.WriteRune(r)
			continue
		}
		return 0, 0, fmt.Errorf("invalid cell reference %s", ref)
	}
	col, err := colLettersToIndex(letters.String())
	if err != nil {
		return 0, 0, err
	}
	row, err := strconv.Atoi(numbers.String())
	if err != nil || row < 1 {
		return 0, 0, fmt.Errorf("invalid cell reference %s", ref)
	}
	return row, col, nil
}

func colLettersToIndex(col string) (int, error) {
	col = strings.ToUpper(strings.TrimSpace(col))
	if col == "" {
		return 0, fmt.Errorf("empty column reference")
	}
	value := 0
	for _, r := range col {
		if r < 'A' || r > 'Z' {
			return 0, fmt.Errorf("invalid column reference %s", col)
		}
		value = value*26 + int(r-'A'+1)
	}
	return value, nil
}

func indexToColLetters(index int) string {
	if index < 1 {
		return ""
	}
	letters := make([]byte, 0, 4)
	for index > 0 {
		index--
		letters = append([]byte{byte('A' + index%26)}, letters...)
		index /= 26
	}
	return string(letters)
}

func trimTrailingEmpty(values []string) []string {
	last := len(values)
	for last > 0 && values[last-1] == "" {
		last--
	}
	return append([]string(nil), values[:last]...)
}

func xmlEscape(value string) string {
	var builder strings.Builder
	if err := xml.EscapeText(&builder, []byte(value)); err != nil {
		return value
	}
	return builder.String()
}

const relsXML = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const stylesXML = `<?xml version="1.0" encoding="UTF-8"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="1"><font><sz val="11"/><name val="Calibri"/></font></fonts>
  <fills count="1"><fill><patternFill patternType="none"/></fill></fills>
  <borders count="1"><border/></borders>
  <cellStyleXfs count="1"><xf/></cellStyleXfs>
  <cellXfs count="1"><xf xfId="0"/></cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
