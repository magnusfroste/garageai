"""GarageAI's LiteLLM callbacks, loaded through `litellm_settings.callbacks` in config.yaml.

Tool schemas without "items": JSON Schema allows an array property without "items" (any
item type), and Claude Code sends tool definitions like that. LiteLLM's token counter
assumes OpenAI's rule that "items" is required and fails with KeyError('items'); the
router then skips its pre-call context check, so a prompt larger than the model's
window reaches the garage and breaks mid-stream instead of getting a clean 400.
Adding "items": {} where it is missing means exactly the same thing to the model, and
lets LiteLLM count the prompt.
"""
from litellm.integrations.custom_logger import CustomLogger


def _add_missing_items(node):
    if isinstance(node, dict):
        if node.get("type") == "array" and "items" not in node:
            node["items"] = {}
        for value in node.values():
            _add_missing_items(value)
    elif isinstance(node, list):
        for value in node:
            _add_missing_items(value)


class GarageAICallbacks(CustomLogger):
    async def async_pre_call_hook(self, user_api_key_dict, cache, data, call_type):
        if data.get("tools"):
            _add_missing_items(data["tools"])
        return data


proxy_handler_instance = GarageAICallbacks()
