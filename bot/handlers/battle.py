from aiogram import Router, F
from aiogram.types import CallbackQuery, Message
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup

import api
from keyboards import battle_result_keyboard, pvp_menu, dungeon_menu, main_menu

router = Router()


class PvPStates(StatesGroup):
    waiting_for_opponent_id = State()


DUNGEON_NAMES = {"easy": "Лёгкий", "medium": "Средний", "hard": "Сложный"}


@router.callback_query(F.data == "pvp")
async def cb_pvp(callback: CallbackQuery):
    user_id = callback.from_user.id
    await callback.message.edit_text(
        f"🏆 <b>PvP Арена</b>\n\n"
        f"Твой ID: <code>{user_id}</code>\n"
        f"Отправь свой ID другу, чтобы он мог бросить тебе вызов!",
        reply_markup=pvp_menu(),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data == "pvp_enter_id")
async def cb_pvp_enter_id(callback: CallbackQuery, state: FSMContext):
    await callback.message.edit_text(
        "🏆 Введи Telegram ID противника:",
        parse_mode="HTML",
    )
    await state.set_state(PvPStates.waiting_for_opponent_id)
    await callback.answer()


@router.message(PvPStates.waiting_for_opponent_id)
async def msg_pvp_opponent_id(message: Message, state: FSMContext):
    await state.clear()

    try:
        defender_id = int(message.text.strip())
    except (ValueError, AttributeError):
        await message.answer(
            "Неверный ID. Введи числовой Telegram ID.",
            reply_markup=main_menu(),
        )
        return

    attacker_id = message.from_user.id
    if attacker_id == defender_id:
        await message.answer(
            "Нельзя сражаться с самим собой!",
            reply_markup=main_menu(),
        )
        return

    try:
        result = await api.battle_pvp(attacker_id, defender_id)
    except Exception as e:
        err = str(e)
        if "deck is empty" in err or "deck" in err.lower():
            await message.answer(
                "Ошибка: у одного из игроков пустая колода!",
                reply_markup=main_menu(),
            )
        else:
            await message.answer(f"Ошибка: {e}", reply_markup=main_menu())
        return

    battle_id = result["battle_id"]
    winner = result["winner"]
    rounds = result["rounds"]

    if winner == "attacker":
        result_text = "🎉 Ты победил!"
    elif winner == "defender":
        result_text = "💀 Ты проиграл..."
    else:
        result_text = "🤝 Ничья!"

    await message.answer(
        f"🏆 <b>PvP Бой</b>\n\n"
        f"{result_text}\n"
        f"Раундов: {rounds}",
        reply_markup=battle_result_keyboard(battle_id),
        parse_mode="HTML",
    )


@router.callback_query(F.data.startswith("pve:"))
async def cb_pve(callback: CallbackQuery):
    dungeon = callback.data.split(":")[1]
    user_id = callback.from_user.id

    try:
        result = await api.battle_pve(user_id, dungeon)
    except Exception as e:
        err = str(e)
        if "deck is empty" in err:
            await callback.answer("Сначала собери колоду!", show_alert=True)
        else:
            await callback.answer(f"Ошибка: {e}", show_alert=True)
        return

    battle_id = result["battle_id"]
    winner = result["winner"]
    rounds = result["rounds"]

    if winner == "attacker":
        result_text = "🎉 Победа!"
    else:
        result_text = "💀 Поражение..."

    dungeon_name = DUNGEON_NAMES.get(dungeon, dungeon)

    await callback.message.edit_text(
        f"⚔️ <b>Данж: {dungeon_name}</b>\n\n"
        f"{result_text}\n"
        f"Раундов: {rounds}",
        reply_markup=battle_result_keyboard(battle_id),
        parse_mode="HTML",
    )
    await callback.answer()


def register_battle_handlers(dp):
    dp.include_router(router)
