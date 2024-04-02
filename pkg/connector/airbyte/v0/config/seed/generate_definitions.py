import jsonref
import json

from urllib.request import urlopen
from os.path import dirname


url = 'https://connectors.airbyte.com/files/registries/v0/oss_registry.json'
response = urlopen(url)
data_json = json.loads(response.read())
definitions = data_json['destinations']
oneOfs = []
vendor_attribute = {}
for idx, _ in enumerate(definitions):
    definitions[idx]['uid'] = definitions[idx]['destinationDefinitionId']
    definitions[idx][
        'id'] = f"airbyte-{definitions[idx]['dockerRepository'].split('/')[1]}"
    definitions[idx]['title'] = "Airbyte " + definitions[idx]['name']
    vendor_attribute[f"airbyte-{definitions[idx]['dockerRepository'].split('/')[1]}"] = definitions[idx]['dockerRepository']+":"+definitions[idx]['dockerImageTag']

    definitions[idx]['spec']['connectionSpecification']['properties']["destination"] = {
        "type": "string",
        "const": definitions[idx]["id"]
    }
    title = definitions[idx]["id"]
    title = title.replace("airbyte-destination-", "")
    title = title.replace("-", "")
    title = title.capitalize()
    definitions[idx]['spec']['connectionSpecification']["title"] = title

    definitions[idx]['spec']['connectionSpecification']['required'].append(
        "destination")
    oneOfs.append(
        definitions[idx]['spec']['connectionSpecification']

    )
    definitions[idx]['spec']['resource_specification'] = definitions[idx]['spec']['connectionSpecification']

new_def = {
    "available_tasks": [
        "TASK_WRITE_DESTINATION"
    ],
    "custom": False,
    "documentation_url": "https://docs.airbyte.com/integrations/destinations",
    "icon": "Airbyte/airbyte.svg",
    "icon_url": "",
    "id": "airbyte-destination",
    "public": True,
    "spec": {
        "resource_specification": {
          "$schema": "http://json-schema.org/draft-07/schema#",
          "title": "Destination",
          "oneOf": oneOfs,
          "type": "object"
        }
    },
    "title": "Airbyte Destination",
    "tombstone": False,
    "type": "CONNECTOR_TYPE_DATA",
    "uid": "975678a2-5117-48a4-a135-019619dee18e",
    "vendor": "Airbyte",
    "vendor_attributes": vendor_attribute
}

new_def = json.dumps(new_def, indent=2, sort_keys=True)
new_def = new_def.replace("airbyte_secret", "instillCredentialField")

with open('../definition.json', 'w') as o:
    o.write(new_def)
