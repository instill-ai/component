## It is not used because there is a problem related to container built with Alpine Linux
from transformers import AutoTokenizer
import json
import sys
import os

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

model = params["model"]
tokenizer = AutoTokenizer.from_pretrained(model,
                                          trust_remote_code=True,
                                          force_download=True)

output = { "toke_count": [0] * len(params["text_chunks"]) }

for i, chunk in enumerate(params["text_chunks"]):
    encoding = tokenizer(chunk)
    output["toke_count"][i] = len(encoding["input_ids"])

print(json.dumps(output))
