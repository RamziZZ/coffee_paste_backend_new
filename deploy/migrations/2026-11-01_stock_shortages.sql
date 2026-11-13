-- Laporan bahan/stok habis dari kasir untuk admin.
-- Server juga menjalankan CREATE TABLE IF NOT EXISTS yang sama saat start, file ini untuk impor manual bila perlu.

CREATE TABLE IF NOT EXISTS `stock_shortages` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(80) NOT NULL,
  `category` varchar(30) NOT NULL,
  `note` varchar(200) NOT NULL DEFAULT '',
  `status` enum('habis','dibeli') NOT NULL DEFAULT 'habis',
  `reported_by` bigint(20) UNSIGNED DEFAULT NULL,
  `reported_by_name` varchar(100) NOT NULL,
  `reported_at` datetime NOT NULL DEFAULT current_timestamp(),
  `purchased_by` bigint(20) UNSIGNED DEFAULT NULL,
  `purchased_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_stock_status` (`status`,`reported_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
