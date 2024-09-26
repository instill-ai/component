from io import BytesIO
import json
import base64
import sys

# TODO: Deal with the import error when running the code in the docker container
# from pdf_to_markdown import PDFTransformer

if __name__ == "__main__":
    json_str   = sys.stdin.buffer.read().decode('utf-8')
    params     = json.loads(json_str)
    pdf_string = params["PDF"]

    decoded_bytes = base64.b64decode(pdf_string)
    pdf_file_obj = BytesIO(decoded_bytes)
    pdf = PDFTransformer(x=pdf_file_obj)
    pages = pdf.raw_pages
    output = {
        "required": len(pages) == 0,
    }
    print(json.dumps(output))
