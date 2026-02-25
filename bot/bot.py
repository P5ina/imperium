import asyncio
import logging

from aiogram import Bot, Dispatcher
from aiogram.fsm.storage.memory import MemoryStorage

from config import BOT_TOKEN
from handlers import register_all_handlers

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


async def main():
    if not BOT_TOKEN:
        logger.error("BOT_TOKEN is not set!")
        return

    bot = Bot(token=BOT_TOKEN)
    dp = Dispatcher(storage=MemoryStorage())

    register_all_handlers(dp)

    logger.info("Imperium bot starting...")
    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
