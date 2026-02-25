from aiogram import Router, F
from aiogram.types import CallbackQuery

import api
from keyboards import RARITY_EMOJI, main_menu, dungeon_menu, battle_result_keyboard

router = Router()

ITEM_NAMES = {
    "bronze_key": "🔑 Бронзовый ключ",
    "silver_key": "🔑 Серебряный ключ",
    "gold_key": "🔑 Золотой ключ",
}


def format_loot_result(result: dict) -> str:
    if result["type"] == "card":
        rarity = result.get("rarity", "common")
        emoji = RARITY_EMOJI.get(rarity, "⚪")
        quality = result.get("quality", 1)
        stars = "⭐" * quality
        return f"{emoji} <b>{result['card_id']}</b> {stars}"
    elif result["type"] == "item":
        name = ITEM_NAMES.get(result.get("item_id", ""), result.get("item_id", "?"))
        return f"🎁 {name}"
    return "???"


@router.callback_query(F.data == "case")
async def cb_case(callback: CallbackQuery):
    user_id = callback.from_user.id

    try:
        data = await api.open_case(user_id)
    except Exception as e:
        await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    results = data.get("results", [])
    lines = ["📦 <b>Открытие кейса...</b>\n"]
    for r in results:
        lines.append(f"  {format_loot_result(r)}")

    await callback.message.edit_text(
        "\n".join(lines),
        reply_markup=main_menu(),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data == "dungeon_menu")
async def cb_dungeon_menu(callback: CallbackQuery):
    user_id = callback.from_user.id

    try:
        items = await api.get_items(user_id)
    except Exception:
        items = []

    keys_text = []
    for item in items:
        name = ITEM_NAMES.get(item["item_type"], item["item_type"])
        keys_text.append(f"{name}: {item['quantity']}")

    text = "⚔️ <b>Данжи</b>\n\n"
    if keys_text:
        text += "Твои ключи:\n" + "\n".join(keys_text) + "\n\n"
    else:
        text += "У тебя нет ключей. Открой кейсы!\n\n"
    text += "Выбери данж:"

    await callback.message.edit_text(
        text,
        reply_markup=dungeon_menu(),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data.startswith("dungeon:"))
async def cb_dungeon(callback: CallbackQuery):
    dungeon = callback.data.split(":")[1]
    user_id = callback.from_user.id
    dungeon_names = {"easy": "Лёгкий", "medium": "Средний", "hard": "Сложный"}

    # Consume key and get loot
    try:
        data = await api.enter_dungeon(user_id, dungeon)
    except Exception as e:
        err = str(e)
        if "not enough keys" in err:
            await callback.answer("Недостаточно ключей!", show_alert=True)
        else:
            await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    loot_results = data.get("results", [])

    # Run PvE battle
    battle_id = None
    battle_winner = None
    battle_rounds = 0
    try:
        battle = await api.battle_pve(user_id, dungeon)
        battle_id = battle.get("battle_id")
        battle_winner = battle.get("winner")
        battle_rounds = battle.get("rounds", 0)
    except Exception as e:
        err = str(e)
        if "deck is empty" in err:
            await callback.answer("⚠️ Сначала собери колоду! Зайди в 🃏 Колода и добавь карты.", show_alert=True)
            return
        # other battle errors — show result without battle info

    lines = [f"⚔️ <b>Данж: {dungeon_names.get(dungeon, dungeon)}</b>\n"]

    if battle_winner == "attacker":
        lines.append("🎉 Победа!\n")
    elif battle_winner == "defender":
        lines.append("💀 Поражение...\n")
    elif battle_winner:
        lines.append("🤝 Ничья!\n")

    if battle_rounds:
        lines.append(f"Раундов: {battle_rounds}\n")

    lines.append("Награда:")
    for r in loot_results:
        lines.append(f"  {format_loot_result(r)}")

    kb = battle_result_keyboard(battle_id) if battle_id else main_menu()

    await callback.message.edit_text(
        "\n".join(lines),
        reply_markup=kb,
        parse_mode="HTML",
    )
    await callback.answer()


def register_loot_handlers(dp):
    dp.include_router(router)
