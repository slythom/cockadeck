PRAGMA foreign_keys=OFF;--> statement-breakpoint
ALTER TABLE `collection_cards` RENAME TO `deck_cards`;--> statement-breakpoint
ALTER TABLE `deck_cards` RENAME COLUMN `collection_id` TO `deck_id`;--> statement-breakpoint
ALTER TABLE `collections` RENAME TO `decks`;--> statement-breakpoint
PRAGMA foreign_keys=ON;
