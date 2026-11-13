-- Harga minuman dua macam: Regular (kolom `price`) dan Large (kolom `price_large`).
-- Server menjalankan langkah yang sama secara otomatis saat start (MenuStore.EnsurePriceSchema),
-- file ini untuk impor manual bila perlu.
--
-- Langkah 1: jalankan SEKALI. Error "Duplicate column name 'price_large'" berarti kolom sudah ada, lewati saja.
ALTER TABLE `menu_items`
  ADD COLUMN `price_large` int(10) UNSIGNED DEFAULT NULL AFTER `price`;

-- Langkah 2: minuman yang sudah ada diberi harga Large = Regular + 4.000 (bisa diubah admin lewat form menu).
-- Makanan dan pastry dibiarkan NULL karena tidak punya pilihan ukuran.
UPDATE `menu_items` m
  JOIN `menu_categories` c ON c.`id` = m.`category_id`
   SET m.`price_large` = m.`price` + 4000
 WHERE c.`is_customizable` = 1 AND m.`price_large` IS NULL;
