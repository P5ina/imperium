from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton, WebAppInfo
from config import MINI_APP_URL

RARITY_EMOJI = {
    "common": "⚪",
    "uncommon": "🟢",
    "rare": "🔵",
    "epic": "🟣",
    "legendary": "🟠",
}


def main_menu():
    return InlineKeyboardMarkup(inline_keyboard=[
        [
            InlineKeyboardButton(text="🎒 Инвентарь", callback_data="inventory:0"),
            InlineKeyboardButton(text="🃏 Колода", callback_data="deck"),
        ],
        [
            InlineKeyboardButton(text="📦 Кейс", callback_data="case"),
            InlineKeyboardButton(text="⚔️ Данж", callback_data="dungeon_menu"),
        ],
        [
            InlineKeyboardButton(text="🏆 PvP", callback_data="pvp"),
        ],
    ])


def inventory_keyboard(page: int, total_pages: int):
    buttons = []
    nav = []
    if page > 0:
        nav.append(InlineKeyboardButton(text="⬅️", callback_data=f"inventory:{page - 1}"))
    nav.append(InlineKeyboardButton(text=f"{page + 1}/{total_pages}", callback_data="noop"))
    if page < total_pages - 1:
        nav.append(InlineKeyboardButton(text="➡️", callback_data=f"inventory:{page + 1}"))
    buttons.append(nav)
    buttons.append([InlineKeyboardButton(text="🔙 Меню", callback_data="main_menu")])
    return InlineKeyboardMarkup(inline_keyboard=buttons)


def deck_keyboard(deck_slots: list):
    buttons = []
    for i in range(1, 6):
        slot = next((s for s in deck_slots if s["slot"] == i), None)
        if slot:
            card = slot["card"]
            defn = card.get("definition", {})
            name = defn.get("name", card.get("card_id", "?"))
            rarity = defn.get("rarity", "common")
            emoji = RARITY_EMOJI.get(rarity, "⚪")
            text = f"Слот {i}: {emoji} {name}"
        else:
            text = f"Слот {i}: пусто"
        buttons.append([InlineKeyboardButton(text=text, callback_data=f"deck_slot:{i}")])
    buttons.append([InlineKeyboardButton(text="🔙 Меню", callback_data="main_menu")])
    return InlineKeyboardMarkup(inline_keyboard=buttons)


def deck_card_picker(cards: list, slot: int, page: int = 0):
    per_page = 5
    total = len(cards)
    total_pages = max(1, (total + per_page - 1) // per_page)
    start = page * per_page
    end = min(start + per_page, total)

    buttons = []
    for card in cards[start:end]:
        defn = card.get("definition", {})
        name = defn.get("name", card.get("card_id", "?"))
        rarity = defn.get("rarity", "common")
        emoji = RARITY_EMOJI.get(rarity, "⚪")
        hp = defn.get("base_hp", 0)
        dmg = defn.get("base_damage", 0)
        quality = card.get("quality", 1)
        stars = "⭐" * quality
        text = f"{emoji} {name} HP:{hp} DMG:{dmg} {stars}"
        buttons.append([InlineKeyboardButton(text=text, callback_data=f"pick_card:{slot}:{card['id']}")])

    nav = []
    if page > 0:
        nav.append(InlineKeyboardButton(text="⬅️", callback_data=f"deck_pick_page:{slot}:{page - 1}"))
    if total_pages > 1:
        nav.append(InlineKeyboardButton(text=f"{page + 1}/{total_pages}", callback_data="noop"))
    if page < total_pages - 1:
        nav.append(InlineKeyboardButton(text="➡️", callback_data=f"deck_pick_page:{slot}:{page + 1}"))
    if nav:
        buttons.append(nav)

    buttons.append([InlineKeyboardButton(text="🔙 Колода", callback_data="deck")])
    return InlineKeyboardMarkup(inline_keyboard=buttons)


def dungeon_menu():
    return InlineKeyboardMarkup(inline_keyboard=[
        [InlineKeyboardButton(text="🟢 Лёгкий (🔑 бронзовый ключ)", callback_data="dungeon:easy")],
        [InlineKeyboardButton(text="🟡 Средний (🔑 серебряный ключ)", callback_data="dungeon:medium")],
        [InlineKeyboardButton(text="🔴 Сложный (🔑 золотой ключ)", callback_data="dungeon:hard")],
        [InlineKeyboardButton(text="🔙 Меню", callback_data="main_menu")],
    ])


def battle_result_keyboard(battle_id: str):
    return InlineKeyboardMarkup(inline_keyboard=[
        [InlineKeyboardButton(
            text="▶️ Смотреть бой",
            web_app=WebAppInfo(url=f"{MINI_APP_URL}?battle_id={battle_id}")
        )],
        [InlineKeyboardButton(text="🔙 Меню", callback_data="main_menu")],
    ])


def pvp_menu():
    return InlineKeyboardMarkup(inline_keyboard=[
        [InlineKeyboardButton(text="⚔️ Ввести ID противника", callback_data="pvp_enter_id")],
        [InlineKeyboardButton(text="🔙 Меню", callback_data="main_menu")],
    ])
