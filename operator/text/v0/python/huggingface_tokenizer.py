from transformers import AutoTokenizer
import json
import sys

json_str = sys.stdin.buffer.read().decode('utf-8')
# Sample input
# {
#   "model": "xxx",
#   "text_chunks": [
#     "Hello, how are you?",
#     "I'm doing well, thank you!"
#   ]
# }
params = json.loads(json_str)

tokenizer = AutoTokenizer.from_pretrained(params["model"])

output = { "toke_count": [] }

for i, chunk in enumerate(params["text_chunks"]):
    encoding = tokenizer(chunk)
    output["toke_count"][i] = len(encoding["input_ids"])

print(json.dumps(output))
