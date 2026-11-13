-- OTP lupa password untuk aplikasi mobile.
-- Tabel ini terpisah dari `password_resets` (tautan token yang dipakai web), jadi alur web tidak terpengaruh.
-- Server juga menjalankan CREATE TABLE IF NOT EXISTS yang sama saat start, file ini untuk impor manual bila perlu.

CREATE TABLE IF NOT EXISTS `password_reset_otps` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) UNSIGNED NOT NULL,
  `code_hash` varchar(255) NOT NULL,
  `attempts` tinyint(3) UNSIGNED NOT NULL DEFAULT 0,
  `expires_at` datetime NOT NULL,
  `used_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `idx_pro_user` (`user_id`),
  CONSTRAINT `fk_pro_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
