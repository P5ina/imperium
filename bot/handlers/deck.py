from aiogram import Router, F
from aiogram.types import CallbackQuery

import api
from keyboards import deck_keyboard, deck_card_picker

router = Router()


@router.callback_query(F.data == "deck")
async def cb_deck(callback: CallbackQuery):
    user_id = callback.from_user.id
    try:
        deck = await api.get_deck(user_id)
    except Exception as e:
        await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    await callback.message.edit_text(
        "🃏 <b>Твоя колода</b>\n\nВыбери слот, чтобы заменить карту:",
        reply_markup=deck_keyboard(deck),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data.startswith("deck_slot:"))
async def cb_deck_slot(callback: CallbackQuery):
    slot = int(callback.data.split(":")[1])
    user_id = callback.from_user.id

    try:
        cards = await api.get_inventory(user_id)
    except Exception as e:
        await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    if not cards:
        await callback.answer("Инвентарь пуст!", show_alert=True)
        return

    await callback.message.edit_text(
        f"🃏 Выбери карту для <b>слота {slot}</b>:",
        reply_markup=deck_card_picker(cards, slot),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data.startswith("deck_pick_page:"))
async def cb_deck_pick_page(callback: CallbackQuery):
    parts = callback.data.split(":")
    slot = int(parts[1])
    page = int(parts[2])
    user_id = callback.from_user.id

    try:
        cards = await api.get_inventory(user_id)
    except Exception as e:
        await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    await callback.message.edit_text(
        f"🃏 Выбери карту для <b>слота {slot}</b>:",
        reply_markup=deck_card_picker(cards, slot, page),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data.startswith("pick_card:"))
async def cb_pick_card(callback: CallbackQuery):
    parts = callback.data.split(":")
    slot = int(parts[1])
    user_card_id = parts[2]
    user_id = callback.from_user.id

    try:
        # Get current deck
        current_deck = await api.get_deck(user_id)
        slots = [{"slot": s["slot"], "user_card_id": s["card"]["id"]} for s in current_deck]

        # Update or add the slot
        found = False
        for s in slots:
            if s["slot"] == slot:
                s["user_card_id"] = user_card_id
                found = True
                break
        if not found:
            slots.append({"slot": slot, "user_card_id": user_card_id})

        await api.set_deck(user_id, slots)
    except Exception as e:
        await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    await callback.answer("Карта установлена!")

    # Refresh deck view
    try:
        deck = await api.get_deck(user_id)
    except Exception:
        deck = []

    await callback.message.edit_text(
        "🃏 <b>Твоя колода</b>\n\nВыбери слот, чтобы заменить карту:",
        reply_markup=deck_keyboard(deck),
        parse_mode="HTML",
    )


def register_deck_handlers(dp):
    dp.include_router(router)
