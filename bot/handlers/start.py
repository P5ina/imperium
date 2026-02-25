from aiogram import Router, F
from aiogram.types import Message, CallbackQuery
from aiogram.filters import CommandStart

import api
from keyboards import main_menu

router = Router()


@router.message(CommandStart())
async def cmd_start(message: Message):
    user = message.from_user
    try:
        await api.register_user(user.id, user.username or "")
    except Exception:
        pass
    await message.answer(
        "🏰 <b>Добро пожаловать в Imperium!</b>\n\n"
        "Собирай карты, собирай колоду и сражайся!",
        reply_markup=main_menu(),
        parse_mode="HTML",
    )


@router.callback_query(F.data == "main_menu")
async def cb_main_menu(callback: CallbackQuery):
    await callback.message.edit_text(
        "🏰 <b>Imperium</b> — Главное меню",
        reply_markup=main_menu(),
        parse_mode="HTML",
    )
    await callback.answer()


@router.callback_query(F.data == "noop")
async def cb_noop(callback: CallbackQuery):
    await callback.answer()


def register_start_handlers(dp):
    dp.include_router(router)
