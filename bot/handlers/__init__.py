from handlers.start import register_start_handlers
from handlers.inventory import register_inventory_handlers
from handlers.deck import register_deck_handlers
from handlers.loot import register_loot_handlers
from handlers.battle import register_battle_handlers


def register_all_handlers(dp):
    register_start_handlers(dp)
    register_inventory_handlers(dp)
    register_deck_handlers(dp)
    register_loot_handlers(dp)
    register_battle_handlers(dp)
