from tokenizers import Tokenizer
import requests
import json
import sys
import re

# Rewrite Tiktoken from Golang package tiktoken with a modification in the attributes
class Tiktoken:
    def __init__(self, tokenizer, special_tokens_set):
        self.tokenizer = tokenizer
        self.special_tokens_set = special_tokens_set

    def difference(self, set1, set2):
        return {token for token in set1 if token not in set2}

    def find_regex_match(self, text, regex):
        match = re.search(regex, text)
        return match.group(0) if match else ""

    def special_token_regex(self, disallowed_special_set):
        special_regex_strs = [re.escape(token) for token in disallowed_special_set]

        special_regex_pattern = "|".join(special_regex_strs)

        special_regex = re.compile(special_regex_pattern)

        return special_regex

    def encode(self, text, allowed_special, disallowed_special):
        if not allowed_special:
            allowed_special_set = set()
        elif len(allowed_special) == 1 and allowed_special[0] == "all":
            allowed_special_set = self.special_tokens_set
        else:
            allowed_special_set = set(allowed_special)

        disallowed_special_set = set(disallowed_special)
        if len(disallowed_special_set) == 1 and "all" in disallowed_special_set:
            disallowed_special_set = self.difference(self.special_tokens_set, allowed_special_set)

        if disallowed_special_set:
            special_regex = self.special_token_regex(disallowed_special_set)
            match = self.find_regex_match(text, special_regex)
            if match:
                raise ValueError(f"text contains disallowed special token '{match}'")

        if len(allowed_special_set) > 0:
            self.tokenizer.add_special_tokens(list(allowed_special_set))
            tokens = self.tokenizer.encode(text, add_special_tokens=True)
        else:
            tokens = self.tokenizer.encode(text, add_special_tokens=False)

        return tokens




json_str = sys.stdin.buffer.read().decode('utf-8')
# params := map[string]interface{}{
# 	"chunk_size":         sp.ChunkSize,
# 	"chunk_overlap":      sp.ChunkOverlap,
# 	"model":              sp.Model,
# 	"allowed_special":    sp.AllowedSpecial,
# 	"disallowed_special": sp.DisallowedSpecial,
# 	"text":               text,
# }

params = json.loads(json_str)
chunk_size = params["chunk_size"]
chunk_overlap = params["chunk_overlap"]
model = params["model"]
allowed_special = params["allowed_special"]
disallowed_special = params["disallowed_special"]
text = params["text"]

url = f"https://api.cohere.com/v1/models/{model}"
headers = {
    "accept": "application/json",
    "authorization": "Bearer ZgfnmYBuFNcFUhYW3xEZeKxwVT6pCfb4YFZ0vUIE"
}

response = requests.get(url, headers=headers)
json_response = json.loads(response.text)
tokenizer_url = json_response["tokenizer_url"]

response = requests.get(tokenizer_url)
tokenizer = Tokenizer.from_str(response.text)
special_tokens = list(json.loads(response.text)["post_processor"]["special_tokens"].keys())

tiktoken = Tiktoken(tokenizer, special_tokens)

tokens = tiktoken.encode(text, allowed_special, disallowed_special)
input_ids = tokens.ids

splits = []
start_idx = 0
cur_idx = min(chunk_size, len(input_ids))

while start_idx < len(input_ids):
    chunk_ids = input_ids[start_idx:cur_idx]
    splits.append(tokenizer.decode(chunk_ids))
    start_idx += chunk_size - chunk_overlap
    cur_idx = min(start_idx + chunk_size, len(input_ids))

output = { "chunks": splits }
print(json.dumps(output))
