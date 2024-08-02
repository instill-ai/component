from tokenizers import Tokenizer
import requests
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


url = f"https://api.cohere.com/v1/models/{params["model"]}"

headers = {
    "accept": "application/json",
    "authorization": "Bearer ZgfnmYBuFNcFUhYW3xEZeKxwVT6pCfb4YFZ0vUIE"
}

response = requests.get(url, headers=headers)
json_response = json.loads(response.text)

tokenizer_url = json_response["tokenizer_url"]

response = requests.get(tokenizer_url)  

tokenizer = Tokenizer.from_str(response.text)

output = { "token_count_map": [] }

for i, chunk in enumerate(params["text_chunks"]):
    result = tokenizer.encode(sequence=chunk, add_special_tokens=False)
    output["token_count_map"][i] = len(result.ids)

print(json.dumps(output))