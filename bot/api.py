import aiohttp
from config import API_URL


async def _request(method: str, path: str, json=None):
    async with aiohttp.ClientSession() as session:
        async with session.request(method, f"{API_URL}{path}", json=json) as resp:
            if resp.status >= 400:
                text = await resp.text()
                raise Exception(f"API error {resp.status}: {text}")
            return await resp.json()


async def register_user(user_id: int, username: str):
    return await _request("POST", "/users", {"id": user_id, "username": username})


async def get_inventory(user_id: int):
    return await _request("GET", f"/users/{user_id}/inventory")


async def get_deck(user_id: int):
    return await _request("GET", f"/users/{user_id}/deck")


async def set_deck(user_id: int, slots: list):
    return await _request("PUT", f"/users/{user_id}/deck", {"slots": slots})


async def get_items(user_id: int):
    return await _request("GET", f"/users/{user_id}/items")


async def open_case(user_id: int):
    return await _request("POST", "/loot/case", {"user_id": user_id})


async def enter_dungeon(user_id: int, dungeon: str):
    return await _request("POST", "/loot/dungeon", {"user_id": user_id, "dungeon": dungeon})


async def battle_pve(user_id: int, dungeon: str):
    return await _request("POST", "/battle/pve", {"user_id": user_id, "dungeon": dungeon})


async def battle_pvp(attacker_id: int, defender_id: int):
    return await _request("POST", "/battle/pvp", {"attacker_id": attacker_id, "defender_id": defender_id})


async def get_battle(battle_id: str):
    return await _request("GET", f"/battle/{battle_id}")
