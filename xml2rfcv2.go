// XML2RFC v2 rendering backend

package mmark

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// XML2 renderer configuration options.
const (
	XML2_STANDALONE = 1 << iota // create standalone document
)

// Xml2 is a type that implements the Renderer interface for XML2RFV2 output.
//
// Do not create this directly, instead use the Xml2Renderer function.
type Xml2 struct {
	flags        int // XML2_* options
	sectionLevel int // current section level
	docLevel     int // frontmatter/mainmatter or backmatter

	// Store the IAL we see for this block element
	ial []*IAL

	// TitleBlock in TOML
	titleBlock *title
}

func (options *Xml2) SetIAL(i []*IAL)        { options.ial = append(options.ial, i...) }
func (options *Xml2) GetAndResetIAL() []*IAL { i := options.ial; options.ial = nil; return i }

// Xml2Renderer creates and configures a Xml2 object, which
// satisfies the Renderer interface.
//
// flags is a set of XML2_* options ORed together
func Xml2Renderer(flags int) Renderer {
	return &Xml2{flags: flags}
}

func (options *Xml2) GetFlags() int {
	return options.flags
}

func (options *Xml2) GetState() int {
	return 0
}

// render code chunks using verbatim, or listings if we have a language
func (options *Xml2) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	s := renderIAL(options.GetAndResetIAL())
	if lang == "" {
		out.WriteString("<sourcecode" + s + ">\n")
	} else {
		out.WriteString("\n<sourcecode" + s + "type=\"" + lang + "\">\n")
	}
	out.Write(text)
	if lang == "" {
		out.WriteString("</sourcecode>\n")
	} else {
		out.WriteString("</sourcecode>\n")
	}
}

func (options *Xml2) TitleBlock(out *bytes.Buffer, text []byte) {}

func (options *Xml2) TitleBlockTOML(out *bytes.Buffer, block *title) {
	if options.flags&XML_STANDALONE == 0 {
		return
	}
	options.titleBlock = block
	out.WriteString("<rfc xmlns:xi=\"http://www.w3.org/2001/XInclude\" ipr=\"" +
		options.titleBlock.Ipr + "\" category=\"" +
		options.titleBlock.Category + "\" docName=\"" + options.titleBlock.DocName + "\">\n")
	out.WriteString("<front>\n")
	out.WriteString("<title abbrev=\"" + options.titleBlock.Abbrev + "\">")
	out.WriteString(options.titleBlock.Title + "</title>\n\n")

	year := ""
	if options.titleBlock.Date.Year() > 0 {
		year = " year=\"" + strconv.Itoa(options.titleBlock.Date.Year()) + "\""
	}
	month := ""
	if options.titleBlock.Date.Month() > 0 {
		month = " month=\"" + time.Month(options.titleBlock.Date.Month()).String() + "\""
	}
	day := ""
	if options.titleBlock.Date.Day() > 0 {
		day = " day=\"" + strconv.Itoa(options.titleBlock.Date.Day()) + "\""
	}
	out.WriteString("<date" + year + month + day + "/>\n\n")

	out.WriteString("<area>" + options.titleBlock.Area + "</area>\n")
	out.WriteString("<workgroup>" + options.titleBlock.Workgroup + "</workgroup>\n")
	for _, k := range options.titleBlock.Keyword {
		out.WriteString("<keyword>" + k + "</keyword>\n")
	}
	for _, a := range options.titleBlock.Author {
		out.WriteString("<author>\n")
		out.WriteString("<initials>" + a.Initials + "</initials>\n")
		out.WriteString("<surname>" + a.Surname + "</surname>\n")
		out.WriteString("<fullname>" + a.Fullname + "</fullname>\n")
		out.WriteString("<role>" + a.Role + "</role>\n")
		out.WriteString("<ascii>" + a.Ascii + "</ascii>\n")
		out.WriteString("</author>\n")
	}
	// Author information
	out.WriteString("\n")
}

func (options *Xml2) BlockQuote(out *bytes.Buffer, text []byte) {
	s := renderIAL(options.GetAndResetIAL())
	out.WriteString("<blockquote" + s + ">\n")
	out.Write(text)
	out.WriteString("</blockquote>\n")
}

func (options *Xml2) Abstract(out *bytes.Buffer, text []byte) {
	out.WriteString("<abstract>\n")
	out.Write(text)
	out.WriteString("</abstract>\n")
}

func (options *Xml2) Aside(out *bytes.Buffer, text []byte) {
	out.WriteString("<aside>\n")
	out.Write(text)
	out.WriteString("</aside>\n")
}

func (options *Xml2) Note(out *bytes.Buffer, text []byte) {
	out.WriteString("<note>\n")
	out.Write(text)
	out.WriteString("</note>\n")
}

func (options *Xml2) Figure(out *bytes.Buffer, text []byte) {
	out.WriteString("<figure>\n")
	out.Write(text)
	out.WriteString("</figure>\n")
}

func (options *Xml2) BlockHtml(out *bytes.Buffer, text []byte) {
	// a pretty lame thing to do...
	out.WriteString("\n\\begin{verbatim}\n")
	out.Write(text)
	out.WriteString("\n\\end{verbatim}\n")
}

func (options *Xml2) Header(out *bytes.Buffer, text func() bool, level int, id string, quote bool) {
	// set amount of open in options, so we know what to close after we finish
	// parsing the doc.
	//marker := out.Len()
	//out.Truncate(marker)
	if quote { // this is a header inside an quoted text block (figure, aside)
		out.WriteString("<name>") // typeset this differently.
		text()
		out.WriteString("</name>\n")
		return
	}

	if level <= options.sectionLevel {
		// close previous ones
		for i := options.sectionLevel - level + 1; i > 0; i-- {
			out.WriteString("</section>\n")
		}
	}
	// new section
	out.WriteString("\n<section anchor=\"" + id + "\">\n")
	out.WriteString("<name>")
	text() // check bool here
	out.WriteString("</name>\n")
	options.sectionLevel = level
	return
}

func (options *Xml2) HRule(out *bytes.Buffer) {
	// not used
}

func (options *Xml2) List(out *bytes.Buffer, text func() bool, flags, start int) {
	marker := out.Len()
	switch {
	case flags&LIST_TYPE_ORDERED != 0:
		if start <= 1 {
			out.WriteString("<ol>\n")
		} else {
			out.WriteString(fmt.Sprintf("<ol start=\"%d\">\n", start))
		}
	case flags&LIST_TYPE_DEFINITION != 0:
		out.WriteString("<dl>\n")
	default:
		out.WriteString("<ul>\n")
	}

	if !text() {
		out.Truncate(marker)
		return
	}
	switch {
	case flags&LIST_TYPE_ORDERED != 0:
		out.WriteString("</ol>\n")
	case flags&LIST_TYPE_DEFINITION != 0:
		out.WriteString("</dl>\n")
	default:
		out.WriteString("</ul>\n")
	}
}

func (options *Xml2) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if flags&LIST_TYPE_DEFINITION != 0 && flags&LIST_TYPE_TERM == 0 {
		out.WriteString("<dd>")
		out.Write(text)
		out.WriteString("</dd>\n")
		return
	}
	if flags&LIST_TYPE_TERM != 0 {
		out.WriteString("<dt>")
		out.Write(text)
		out.WriteString("</dt>\n")
		return
	}
	out.WriteString("<li>")
	out.Write(text)
	out.WriteString("</li>\n")
}

func (options *Xml2) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()
	out.WriteString("<t>")
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("</t>\n")
}

func (options *Xml2) Tables(out *bytes.Buffer, text []byte) {}

func (options *Xml2) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int, table bool) {
	out.WriteString("<table>\n<thead>\n")
	out.Write(header)
	out.WriteString("</thead>\n")
	out.Write(body)
	out.WriteString("</table>\n")
}

func (options *Xml2) TableRow(out *bytes.Buffer, text []byte) {
	out.WriteString("<tr>")
	out.Write(text)
	out.WriteString("</tr>\n")
}

func (options *Xml2) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	a := ""
	switch align {
	case TABLE_ALIGNMENT_LEFT:
		a = " align=\"left\""
	case TABLE_ALIGNMENT_RIGHT:
		a = " align=\"right\""
	default:
		a = " align=\"center\""
	}
	out.WriteString("<th" + a + ">")
	out.Write(text)
	out.WriteString("</th>")

}

func (options *Xml2) TableCell(out *bytes.Buffer, text []byte, align int) {
	out.WriteString("<td>")
	out.Write(text)
	out.WriteString("</td>")
}

func (options *Xml2) Footnotes(out *bytes.Buffer, text func() bool) {
	// not used
}

func (options *Xml2) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	// not used
}

func (options *Xml2) Index(out *bytes.Buffer, primary, secondary []byte) {
	out.WriteString("<iref item=\"" + string(primary) + "\"")
	out.WriteString(" subitem=\"" + string(secondary) + "\"" + "/>")
}

func (options *Xml2) Citation(out *bytes.Buffer, link, title []byte) {
	out.WriteString("<xref target=\"" + string(link) + "\"/>")
}

func (options *Xml2) References(out *bytes.Buffer, citations map[string]*citation, first bool) {
	if !first || options.flags&XML_STANDALONE == 0 {
		return
	}
	// close any option section tags
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch options.docLevel {
	case DOC_FRONT_MATTER:
		out.WriteString("</front>\n")
		out.WriteString("<back>\n")
	case DOC_MAIN_MATTER:
		out.WriteString("</middle>\n")
		out.WriteString("<back>\n")
	case DOC_BACK_MATTER:
		// nothing to do
	}
	options.docLevel = DOC_BACK_MATTER
	// count the references
	refi, refn := 0, 0
	for _, c := range citations {
		if c.typ == 'i' {
			refi++
		}
		if c.typ == 'n' {
			refn++
		}
	}
	// output <xi:include href="<references file>.xml"/>, we use file it its not empty, otherwise
	// we construct one for RFCNNNN and I-D.something something.
	if refi+refn > 0 {
		if refi > 0 {
			out.WriteString("<references title=\"Informative References\">\n")
			for _, c := range citations {
				if c.typ == 'i' {
					f := string(c.filename)
					if f == "" {
						f = referenceFile(c)
					}
					out.WriteString("\t<xi:include href=\"" + f + "\"/>\n")
				}
			}
			out.WriteString("</references>\n")
		}
		if refn > 0 {
			out.WriteString("<references title=\"Normative References\">\n")
			for _, c := range citations {
				if c.typ == 'n' {
					f := string(c.filename)
					if f == "" {
						f = referenceFile(c)
					}
					out.WriteString("\t<xi:include href=\"" + f + "\"/>\n")
				}
			}
			out.WriteString("</references>\n")
		}
	}
}

func (options *Xml2) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.WriteString("\\href{")
	if kind == LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}
	out.Write(link)
	out.WriteString("}{")
	out.Write(link)
	out.WriteString("}")
}

func (options *Xml2) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("<tt>")
	convertEntity(out, text)
	out.WriteString("</tt>")
}

func (options *Xml2) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	// Check for 2119 Keywords
	s := string(text)
	if _, ok := words2119[s]; ok {
		out.WriteString("<bcp14>")
		out.Write(text)
		out.WriteString("</bcp14>")
		return
	}
	out.WriteString("<strong>")
	out.Write(text)
	out.WriteString("</strong>")
}

func (options *Xml2) Emphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<em>")
	out.Write(text)
	out.WriteString("</em>")
}

func (options *Xml2) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	if bytes.HasPrefix(link, []byte("http://")) || bytes.HasPrefix(link, []byte("https://")) {
		// treat it like a link
		out.WriteString("\\href{")
		out.Write(link)
		out.WriteString("}{")
		out.Write(alt)
		out.WriteString("}")
	} else {
		out.WriteString("\\includegraphics{")
		out.Write(link)
		out.WriteString("}")
	}
}

func (options *Xml2) LineBreak(out *bytes.Buffer) {
	out.WriteString("\n<vspace/>\n")
}

func (options *Xml2) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteString("\\href{")
	out.Write(link)
	out.WriteString("}{")
	out.Write(content)
	out.WriteString("}")
}

func (options *Xml2) RawHtmlTag(out *bytes.Buffer, tag []byte) {
}

func (options *Xml2) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong><em>")
	out.Write(text)
	out.WriteString("</em></strong>")
}

func (options *Xml2) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (options *Xml2) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	// not used
}

func (options *Xml2) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (options *Xml2) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

// header and footer
func (options *Xml2) DocumentHeader(out *bytes.Buffer, first bool) {
	if !first || options.flags&XML_STANDALONE == 0 {
		return
	}
	out.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
}

func (options *Xml2) DocumentFooter(out *bytes.Buffer, first bool) {
	if !first || options.flags&XML_STANDALONE == 0 {
		return
	}
	// close any option section tags
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch options.docLevel {
	case DOC_FRONT_MATTER:
		out.WriteString("</front>\n")
	case DOC_MAIN_MATTER:
		out.WriteString("</middle>\n")
	case DOC_BACK_MATTER:
		out.WriteString("</back>\n")
	}
	out.WriteString("</rfc>\n")
}

func (options *Xml2) DocumentMatter(out *bytes.Buffer, matter int) {
	// we default to frontmatter already openened in the documentHeader
	switch matter {
	case DOC_FRONT_MATTER:
		// already open
	case DOC_MAIN_MATTER:
		out.WriteString("</front>\n")
		out.WriteString("<middle>\n")
	case DOC_BACK_MATTER:
		out.WriteString("</middle>\n")
		out.WriteString("<back>\n")
	}
	options.docLevel = matter
}