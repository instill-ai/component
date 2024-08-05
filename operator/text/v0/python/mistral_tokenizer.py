import json
import sys
from mistral_common.tokens.tokenizers.mistral import MistralTokenizer
from mistral_common.protocol.instruct.request import ChatCompletionRequest
from mistral_common.protocol.instruct.messages import UserMessage

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

tokenizer = MistralTokenizer.from_model(params["model"])

output = { "toke_count": [0] * len(params["text_chunks"]) }

for i, chunk in enumerate(params["text_chunks"]):
    res = tokenizer.encode_chat_completion(
        ChatCompletionRequest(messages=[UserMessage(content=chunk)])
    )
    output["toke_count"][i] = len(res.tokens)

print(json.dumps(output))
