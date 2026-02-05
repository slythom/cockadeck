CREATE TABLE `card` (
	`id` int AUTO_INCREMENT NOT NULL,
	`scryfall_id` varchar(36) NOT NULL,
	`name` varchar(255) NOT NULL,
	`set_code` varchar(10) NOT NULL,
	`collector_number` varchar(20) NOT NULL,
	`image_url` text NOT NULL,
	`oracle_text` text,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `card_id` PRIMARY KEY(`id`),
	CONSTRAINT `card_scryfall_id_unique` UNIQUE(`scryfall_id`)
);
--> statement-breakpoint
CREATE TABLE `collection_card` (
	`user_id` varchar(36) NOT NULL,
	`card_id` int NOT NULL,
	`quantity` int NOT NULL DEFAULT 1,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `collection_card_user_id_card_id_pk` PRIMARY KEY(`user_id`,`card_id`)
);
--> statement-breakpoint
CREATE TABLE `deck` (
	`id` int AUTO_INCREMENT NOT NULL,
	`user_id` varchar(36) NOT NULL,
	`name` varchar(255) NOT NULL,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	`updated_at` timestamp NOT NULL DEFAULT (now()) ON UPDATE CURRENT_TIMESTAMP,
	CONSTRAINT `deck_id` PRIMARY KEY(`id`),
	CONSTRAINT `deck_user_name_unique` UNIQUE(`user_id`,`name`)
);
--> statement-breakpoint
CREATE TABLE `deck_card` (
	`deck_id` int NOT NULL,
	`card_id` int NOT NULL,
	`quantity` int NOT NULL DEFAULT 1,
	CONSTRAINT `deck_card_deck_id_card_id_pk` PRIMARY KEY(`deck_id`,`card_id`)
);
--> statement-breakpoint
CREATE TABLE `session` (
	`id` varchar(255) NOT NULL,
	`user_id` varchar(36) NOT NULL,
	`expires_at` timestamp NOT NULL,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `session_id` PRIMARY KEY(`id`)
);
--> statement-breakpoint
CREATE TABLE `user` (
	`id` varchar(36) NOT NULL,
	`email` varchar(255) NOT NULL,
	`password_hash` varchar(255) NOT NULL,
	`created_at` timestamp NOT NULL DEFAULT (now()),
	CONSTRAINT `user_id` PRIMARY KEY(`id`),
	CONSTRAINT `user_email_unique` UNIQUE(`email`)
);
--> statement-breakpoint
ALTER TABLE `collection_card` ADD CONSTRAINT `collection_card_user_id_user_id_fk` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE `collection_card` ADD CONSTRAINT `collection_card_card_id_card_id_fk` FOREIGN KEY (`card_id`) REFERENCES `card`(`id`) ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE `deck` ADD CONSTRAINT `deck_user_id_user_id_fk` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE `deck_card` ADD CONSTRAINT `deck_card_deck_id_deck_id_fk` FOREIGN KEY (`deck_id`) REFERENCES `deck`(`id`) ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE `deck_card` ADD CONSTRAINT `deck_card_card_id_card_id_fk` FOREIGN KEY (`card_id`) REFERENCES `card`(`id`) ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE `session` ADD CONSTRAINT `session_user_id_user_id_fk` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX `card_search_idx` ON `card` (`name`,`set_code`,`collector_number`);--> statement-breakpoint
CREATE INDEX `collection_card_card_idx` ON `collection_card` (`card_id`);--> statement-breakpoint
CREATE INDEX `deck_user_idx` ON `deck` (`user_id`);--> statement-breakpoint
CREATE INDEX `deck_card_card_idx` ON `deck_card` (`card_id`);--> statement-breakpoint
CREATE INDEX `session_user_idx` ON `session` (`user_id`);--> statement-breakpoint
CREATE INDEX `session_expires_idx` ON `session` (`expires_at`);