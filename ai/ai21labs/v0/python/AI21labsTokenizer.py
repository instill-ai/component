from argparse import ArgumentParser
from ai21_tokenizer import Tokenizer, PreTrainedTokenizers
from enum import Enum
import json


class AI21labsModel(str, Enum):
    mini = "jamba-1.5-mini"
    large = "jamba-1.5-large"
    instruct = "jamba-instruct-preview"


def create_parser():
    parser = ArgumentParser(description='AI21labs Tokenizer')
    parser.add_argument('--text', type=str, required=True,
                        help='The text to tokenize')
    parser.add_argument('--model', type=AI21labsModel,
                        choices=list(AI21labsModel), default=AI21labsModel.large)
    return parser


def cli():
    p = create_parser()
    args = p.parse_args()
    return args.text, args.model


def main():
    text, model = cli()

    if model == AI21labsModel.mini:
        tokenizer = Tokenizer.get_tokenizer(
            PreTrainedTokenizers.JAMBA_1_5_MINI_TOKENIZER)

    elif model == AI21labsModel.large:
        tokenizer = Tokenizer.get_tokenizer(
            PreTrainedTokenizers.JAMBA_1_5_LARGE_TOKENIZER)

    elif model == AI21labsModel.instruct:
        tokenizer = Tokenizer.get_tokenizer(
            PreTrainedTokenizers.JAMBA_INSTRUCT_TOKENIZER)

    else:
        raise ValueError(f"Unknown model: {model}")

    encoded_text = tokenizer.encode(text)

    result = {
        "text": text,
        "model": model,
        "token_count": len(encoded_text),
    }

    print(json.dumps(result))


if __name__ == '__main__':
    main()
