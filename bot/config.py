import os
from dotenv import load_dotenv

load_dotenv()

BOT_TOKEN = os.getenv("BOT_TOKEN", "")
API_URL = os.getenv("API_URL", "http://localhost:8090")
MINI_APP_URL = os.getenv("MINI_APP_URL", "https://imperium.p5ina.dev")
