from aiogram import Router, F
from aiogram.types import CallbackQuery

import api
from keyboards import RARITY_EMOJI, inventory_keyboard

router = Router()
PER_PAGE = 5


@router.callback_query(F.data.startswith("inventory:"))
async def cb_inventory(callback: CallbackQuery):
    page = int(callback.data.split(":")[1])
    user_id = callback.from_user.id

    try:
        cards = await api.get_inventory(user_id)
    except Exception as e:
        await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    if not cards:
        await callback.message.edit_text(
            "🎒 <b>Инвентарь пуст</b>\n\nОткрой кейс, чтобы получить карты!",
            reply_markup=inventory_keyboard(0, 1),
            parse_mode="HTML",
        )
        await callback.answer()
        return

    total_pages = max(1, (len(cards) + PER_PAGE - 1) // PER_PAGE)
    page = min(page, total_pages - 1)
    start = page * PER_PAGE
    end = min(start + PER_PAGE, len(cards))

    lines = ["🎒 <b>Инвентарь</b>\n"]
    for card in cards[start:end]:
        defn = card.get("definition", {})
        name = defn.get("name", card.get("card_id", "?"))
        rarity = defn.get("rarity", "common")
        emoji = RARITY_EMOJI.get(rarity, "⚪")
        hp = defn.get("base_hp", 0)
        dmg = defn.get("base_damage", 0)
        quality = card.get("quality", 1)
        stars = "⭐" * quality
        lines.append(f"{emoji} <b>{name}</b> — HP:{hp} DMG:{dmg} {stars}")

    await callback.message.edit_text(
        "\n".join(lines),
        reply_markup=inventory_keyboard(page, total_pages),
        parse_mode="HTML",
    )
    await callback.answer()


def register_inventory_handlers(dp):
    dp.include_router(router)
