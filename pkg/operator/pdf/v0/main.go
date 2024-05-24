//go:generate compogen readme --operator ./config ./README.mdx
package pdf

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	_ "embed"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

const (
	taskConvertToMarkdown string = "TASK_CONVERT_TO_MARKDOWN"
	scriptPath            string = "/component/pkg/operator/pdf/v0/python/pdf_transformer.py"
	pythonInterpreter     string = "/opt/venv/bin/python"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte
	once      sync.Once
	op        *operator
)

type operator struct {
	base.Operator
}

type execution struct {
	base.OperatorExecution
}

func Init(bo base.Operator) *operator {
	once.Do(func() {
		op = &operator{Operator: bo}
		err := op.LoadOperatorDefinition(definitionJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})
	return op
}

func (o *operator) CreateExecution(sysVars map[string]any, task string) (*base.ExecutionWrapper, error) {
	return &base.ExecutionWrapper{Execution: &execution{
		OperatorExecution: base.OperatorExecution{Operator: o, SystemVariables: sysVars, Task: task},
	}}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}

	for _, input := range inputs {
		switch e.Task {
		case taskConvertToMarkdown:
			inputStruct := convertPDFToMarkdownInput{}
			err := base.ConvertFromStructpb(input, &inputStruct)
			if err != nil {
				return nil, err
			}

			cmd := exec.Command(pythonInterpreter, "-c", pythonCode)

			outputStruct, err := convertPDFToMarkdown(inputStruct, cmd)
			if err != nil {
				return nil, err
			}
			output, err := base.ConvertToStructpb(outputStruct)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, output)
		default:
			return nil, fmt.Errorf("not supported task: %s", e.Task)
		}
	}
	return outputs, nil
}

// compiled code cannot find the proper path. So, we extract the Python code from python/pdf_transformer.py and put it here.
const (
	pythonCode = `
import pdfplumber
import sys
from io import BytesIO
import json

class PdfTransformer:
	def __init__(self, x):
		# self.path = path
		# x can be a path or a file object.
		self.pdf = pdfplumber.open(x)
		self.pages = self.pdf.pages
		self.metadata = self.pdf.metadata

		self.set_heights()
		self.lines = []
		self.tables = []
		self.images = []
		self.process_image()
		for page in self.pages:
			page_lines = page.extract_text_lines()
			self.process_line(page_lines, page.page_number)
			self.process_table(page)
				
		self.result = ""

	def process_image(self):
		i = 0
		for page in self.pages:
			images = page.images
			for image in images:
				image["page_number"] = page.page_number
				image["img_number"] = i
				i += 1
				self.images.append(image)

	def set_heights(self):
		tolerance = 0.95
		largest_text_height, second_largest_text_height = 0, 0
		for page in self.pages:
			for line in page.extract_text_lines():
				height = int(line["bottom"] - line["top"])
				if height > largest_text_height:
					second_largest_text_height = largest_text_height
					largest_text_height = height
				elif height > second_largest_text_height and height < largest_text_height:
					second_largest_text_height = height
		self.title_height = largest_text_height * tolerance
		self.subtitle_height = second_largest_text_height * tolerance


	def execute(self):
		self.set_line_type(self.title_height, self.subtitle_height, "indent")
		self.result = self.transform_line_to_markdown(self.lines)
		return self.result
		
		# export .md file
		# file_name = "pdfTransform/output/" + self.path.split("/")[-1].replace(".pdf", ".md")
		# with open(file_name, "w") as f:
		#     f.write(self.result)

	# It can add more calculation for the future development when we want to extend more use cases.
	def process_line(self, lines, page_number):
		for idx, line in enumerate(lines):
			line["line_height"] = line["bottom"] - line["top"]
			line["line_width"] = line["x1"] - line["x0"]
			line["middle"] = (line["x1"] + line["x0"]) / 2
			line["distance_to_next_line"] = lines[idx+1]["top"] - line["bottom"] if idx < len(lines) - 1 else None
			line["page_number"] = page_number
			self.lines.append(line)

	def process_table(self, page):
		tables = page.find_tables()
		if tables:
			for table in tables:
				table_info = {}
				table_info["bbox"] = table.bbox
				text = table.extract()
				table_info["text"] = text
				table_info["page_number"] = page.page_number
				self.tables.append(table_info)

	# TODO: Chinese version is not working for bold.
	def is_bold(self, char):
		return char['fontname'] and 'bold' in char['fontname'].lower()

	# TODO: Implement paragraph strategy
	def paragraph_strategy(self, lines, subtitle_height=14):
		# TODO: Implement paragraph strategy
		# judge the non-title line in a page.
		# If there is a line with indent, return "indent"
		# If there is a line with no indent, return "no-indent"
		return "indent"
		paragraph_lines_start_positions = []
		for line in lines:
			if line["line_height"] < subtitle_height:
				paragraph_lines_start_positions.append(line["x0"])

	def set_line_type(self, title_height=16, subtitle_height=14, paragraph_strategy="indent"):
		lines = self.lines
		current_paragraph = []
		paragraph_start_position = 0
		paragraph_idx = 1

		for i, line in enumerate(lines):
			if line['line_height'] >= title_height:
				line["type"] = 'title'
				if current_paragraph:
					for line_in_paragraph in current_paragraph:
						line_in_paragraph["type"] = f'paragraph {paragraph_idx}'       
					paragraph_idx += 1
					current_paragraph = []

			elif line['line_height'] >= subtitle_height:
				line["type"] = 'subtitle'
				if current_paragraph:
					for line_in_paragraph in current_paragraph:
						line_in_paragraph["type"] = f'paragraph {paragraph_idx}'       
					paragraph_idx += 1
					current_paragraph = []
			else:
				if current_paragraph:
					current_paragraph.append(line)
					
					if ((paragraph_strategy == "indent" and i < len(lines) - 1 and 
							(   # if the next line starts a new paragraph
								abs(lines[i+1]['x0'] - paragraph_start_position) < 10 
								# if the next line is not in the same layer
								# or abs(line["middle"] - lines[i+1]["middle"]) > 5
								)
							) or
						(paragraph_strategy == "no-indent" 
							and line["distance_to_next_line"] 
							and line["distance_to_next_line"] > 10) or
						(i == len(lines) - 1) # final line
						):
						
						for line_in_paragraph in current_paragraph:
							line_in_paragraph["type"] = f'paragraph {paragraph_idx}'       
						
						paragraph_idx += 1
						current_paragraph = []
				else:
					current_paragraph = [line]
					paragraph_start_position = line["x0"]
		self.lines = lines

	def transform_line_to_markdown(self, lines):
		result = ""
		to_be_processed_table = []
		for i, line in enumerate(lines):
			table = self.meet_table(line, line["page_number"])
			if table and table not in to_be_processed_table:
				to_be_processed_table.append(table)
			elif table and table in to_be_processed_table:
				continue
			elif to_be_processed_table:
				for table in to_be_processed_table:
					result += "\n\n"
					result += self.transform_table_markdown(table)
					result += "\n\n"
					self.tables.remove(table)
				to_be_processed_table = []
				
				result += self.line_process(line, i, lines, result)
				result += "\n\n"

			else:
				result += self.line_process(line, i, lines, result)
			
			if i < len(lines) - 1:

				result += self.insert_image(line, lines[i+1]) 
			else:
				result += self.insert_image(line, None)
		if self.tables:
			processed_table = []
			for table in self.tables:
				result += "\n\n"
				result += self.transform_table_markdown(table)
				result += "\n\n"
				processed_table.append(table)
			for table in processed_table:
				self.tables.remove(table)
				
		return result

	def line_process(self, line, i, lines, current_result):
		result = ""
		if "type" not in line:
			return line["text"]
		if line["type"] == "title":
			if current_result != "":
				result += "\n"
				result += "\n"
			result += f"# {line['text']}\n"
		elif line["type"] == "subtitle":
			if current_result != "":
				result += "\n"
				result += "\n"
			result += f"## {line['text']}\n"
		elif "paragraph" in line["type"]:
			bold_trigger = False
			bold_text = ""
			## TODO: English version is not working for set the whitespace between words.
			## It can be solved by using extract_words() instead of extract_text_lines()
			## TODO: Chinese version is not working for bold.
			## It is still under investigation.
			for char in line["chars"]:                                    
				# if self.is_bold(char) and not bold_trigger: # start of bold text
				#     bold_text += f"**{char['text']}"
				#     bold_trigger = True
				# elif self.is_bold(char) and bold_trigger: # continue bold text
				#     bold_text += char["text"]
				# elif not self.is_bold(char) and bold_trigger: # end of bold text
				#     bold_text += f"{char['text']}**"
				#     result += bold_text
				#     bold_text = ""
				#     bold_trigger = False
				# else:
				#     result += char["text"]
				result += char["text"]
			# extract numbers from line["type"] and add a change the line when the next line is a new paragraph
			
			if (
				(i < len(lines) - 1) and
				"type" in lines[i+1] and
				len(lines[i+1]["type"].split(" ")) == 2 and
				(int(line["type"].split(" ")[1]) < int(lines[i+1]["type"].split(" ")[1]))
			):
				result += "\n"
				result += "\n"
		return result

	def meet_table(self, line, page_number):
		tables = self.tables
		for table in tables:
			if table["page_number"] == page_number:
				bbox = table["bbox"]
				top, bottom = bbox[1], bbox[3]
				if line["top"] > top and line["bottom"] < bottom:
					return table
				else:
					None

	def transform_table_markdown(self, table):
		result = ""
		texts = table["text"]
		for i, row in enumerate(texts):
			for j, col in enumerate(row):
				if col:
					if "\n" in col:
						col = col.replace("\n", "<br>")
					result += col
					
					if j < len(row) - 1:
						result += " | "
			if i == 0:
				result += "\n"
				## TODO: Judge table that cross the page, 
				result += "|"
				result += " --- |" * len(row)
				result += "\n"
			elif i < len(texts) - 1:
				result += "\n"
			
		return result

	
	def insert_image(self, line, next_line):
		result = ""
		images = self.images
		to_be_removed_images = []
		
		if images:
			if next_line:
				# If there is image between line and next_line, we insert image.
				if next_line["page_number"] == line["page_number"]:
					for image in images:
						if image["page_number"] == line["page_number"] and image["top"] > line["bottom"] and image["bottom"] < next_line["top"]:
							result += "\n\n"
							result += f"![image {image['img_number']}]"
							result += "\n\n"
							to_be_removed_images.append(image)
				elif next_line["page_number"] > line["page_number"]:
					for image in images:
						if image["page_number"] >= line["page_number"] and image["page_number"] < next_line["page_number"]:
							result += "\n\n"
							result += f"![image {image['img_number']}]"
							result += "\n\n"
							to_be_removed_images.append(image)
				
			else: # if images exists and there is no next_line, we insert image.
				for image in images:
					result += "\n\n"
					result += f"![image {image['img_number']}]"
					result += "\n\n"
					to_be_removed_images.append(image)
		for image in to_be_removed_images:
			self.images.remove(image)
		
		return result


if __name__ == "__main__":
	pdf_bytes = sys.stdin.buffer.read()
	pdf_file_obj = BytesIO(pdf_bytes)
	pdf = PdfTransformer(pdf_file_obj)
	result = pdf.execute()
	output = {
		"body": result,
		"metadata": pdf.metadata
	}
	print(json.dumps(output))
	`
)
