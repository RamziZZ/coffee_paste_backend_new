-- MySQL dump 10.13  Distrib 8.0.46, for Win64 (x86_64)
--
-- Host: sakura.proxy.rlwy.net    Database: railway
-- ------------------------------------------------------
-- Server version	8.4.11

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `device_tokens`
--

DROP TABLE IF EXISTS `device_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `device_tokens` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `platform` enum('android','ios') COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_device_token` (`token`),
  KEY `idx_dt_user` (`user_id`),
  CONSTRAINT `fk_dt_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `device_tokens`
--

LOCK TABLES `device_tokens` WRITE;
/*!40000 ALTER TABLE `device_tokens` DISABLE KEYS */;
INSERT INTO `device_tokens` VALUES (17,4,'f-dGRxFuRMySw-ybzbt5ue:APA91bE_pUC3y2h_AsLUG_c8CZktE6f6Zh6Jg7RzqUR3OGLvJcbQCajX1ZAJdF6OO3ZOc07LS_NrapB3XkgelkMzRge9qkKstp_JfEq9RQPRCBYBMQW4LOg','android','2026-10-29 23:14:18'),(18,11,'fj9-VBnhQ3Cf9wdvxNRGT4:APA91bEzETX5TSWiU8zeRWYTShcXTNp3CL6wha1kbe80PCrQy0ccOhaR5AG2YO4gPEieb1unwIGjCqHHQiMhBotcdv_9InVUI7Lfsc5hkZ1gP5k4MSlwne0','android','2026-10-30 09:30:48'),(31,3,'ct6cTyqHSdudDDGcAgbIjr:APA91bGJw7zVfsU2DCJnaLKZDuBTTyDUZdsH3M3kx1HjdsJQnsOk9LJ0qXoIKBBZaNrzUNrGh-m5sUFSFL_ap0b6yIJSFE742Bz_zOI-BZxtJuq0RPd7gos','android','2026-11-01 19:18:40'),(37,6,'d19x0WjpQFWbWmBDWXy8sp:APA91bFfCuoHJz9j-F3ya8E7xNDdQRWVVxOYj4G0loRb_VwIGVWVWW1Tmw02O0tyEDt97GJ7y3kS3N-oiYyvRTAMrKWdOKT4cTgwZwSVyJk0Oy-iNBLgV6A','android','2026-11-03 10:13:23'),(40,13,'cTPvkZ_jRNemPStpCzehG4:APA91bE3xJ6QI-XHgaz_cB6n_z35T-yoi4Z_M5ePP8yV02CrrMG4Bqe8kkUgQ-ileOaWFbKH7r-2VYLqUzLRzfESjI8JkWFF7cE2bkGpEL9lY9slclkRIfs','android','2026-11-03 11:23:36');
/*!40000 ALTER TABLE `device_tokens` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `email_change_requests`
--

DROP TABLE IF EXISTS `email_change_requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `email_change_requests` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `new_email` varchar(150) COLLATE utf8mb4_unicode_ci NOT NULL,
  `token_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `used_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_ecr_user` (`user_id`),
  CONSTRAINT `fk_ecr_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `email_change_requests`
--

LOCK TABLES `email_change_requests` WRITE;
/*!40000 ALTER TABLE `email_change_requests` DISABLE KEYS */;
INSERT INTO `email_change_requests` VALUES (3,2,'ziegoramaa@gmail.com','e41348f851f6e2878764711ec7e4646956bee8147224fcf4bd19918129ad774d','2026-10-09 12:52:38','2026-10-09 12:38:38','2026-10-09 12:37:38'),(4,2,'shintamaria293@gmail.com','6fdb5feb3fa0416d7db830268f99f312aafc9248277185afd347fafcc7d0cca4','2026-10-09 12:54:23','2026-10-09 12:39:53','2026-10-09 12:39:23'),(5,3,'shimar2901@gmail.com','74fa1bc20971c3676c3edab23e870f3853170dff44024c62d981dd3f88f11bbc','2026-10-11 06:57:24','2026-10-11 06:42:51','2026-10-11 06:42:24');
/*!40000 ALTER TABLE `email_change_requests` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `expenses`
--

DROP TABLE IF EXISTS `expenses`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `expenses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `name` varchar(150) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category` enum('bahan-baku','operasional','lainnya') COLLATE utf8mb4_unicode_ci NOT NULL,
  `amount` int unsigned NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_expenses_date` (`date`),
  KEY `fk_expenses_user` (`created_by`),
  CONSTRAINT `fk_expenses_user` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `expenses`
--

LOCK TABLES `expenses` WRITE;
/*!40000 ALTER TABLE `expenses` DISABLE KEYS */;
INSERT INTO `expenses` VALUES (1,'2026-10-05','bayar listrik','operasional',100000,'',1,'2026-10-05 09:49:44','2026-10-05 09:49:44'),(2,'2026-10-07','beli kopi','bahan-baku',300000,'',1,'2026-10-07 10:50:46','2026-10-07 10:50:46'),(3,'2026-10-08','beli coklat','bahan-baku',50000,'',1,'2026-10-08 20:41:03','2026-10-08 20:41:03'),(4,'2026-10-12','gaji karyawan','lainnya',1000000,'',1,'2026-10-12 11:56:24','2026-10-12 11:56:24'),(5,'2026-10-29','bayar listrik','operasional',500000,'',1,'2026-10-29 18:05:06','2026-10-29 18:05:06'),(6,'2026-11-03','biji kopi','bahan-baku',430000,'',1,'2026-11-03 12:36:00','2026-11-03 12:36:00');
/*!40000 ALTER TABLE `expenses` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `favorite_menu`
--

DROP TABLE IF EXISTS `favorite_menu`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `favorite_menu` (
  `user_id` bigint unsigned NOT NULL,
  `menu_item_id` bigint unsigned NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`,`menu_item_id`),
  KEY `idx_fav_menu` (`menu_item_id`),
  CONSTRAINT `fk_fav_menu` FOREIGN KEY (`menu_item_id`) REFERENCES `menu_items` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_fav_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `favorite_menu`
--

LOCK TABLES `favorite_menu` WRITE;
/*!40000 ALTER TABLE `favorite_menu` DISABLE KEYS */;
/*!40000 ALTER TABLE `favorite_menu` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `menu_categories`
--

DROP TABLE IF EXISTS `menu_categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `menu_categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `filter_label` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `icon` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_customizable` tinyint(1) NOT NULL DEFAULT '0',
  `sort_order` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_menu_categories_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `menu_categories`
--

LOCK TABLES `menu_categories` WRITE;
/*!40000 ALTER TABLE `menu_categories` DISABLE KEYS */;
INSERT INTO `menu_categories` VALUES (1,'Coffee','Kopi','coffee-cup',1,1),(2,'Non-Coffee','Non Kopi','coffee-cup',1,2),(3,'Makanan','Makanan','utensils',0,3),(4,'Pastry','Pastry','gift',0,4);
/*!40000 ALTER TABLE `menu_categories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `menu_items`
--

DROP TABLE IF EXISTS `menu_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `menu_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `category_id` bigint unsigned NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `price` int unsigned NOT NULL,
  `price_large` int unsigned DEFAULT NULL,
  `status` enum('tersedia','habis') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'tersedia',
  `image_url` text COLLATE utf8mb4_unicode_ci,
  `is_featured` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_menu_category` (`category_id`),
  KEY `idx_menu_status` (`status`),
  CONSTRAINT `fk_menu_category` FOREIGN KEY (`category_id`) REFERENCES `menu_categories` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `menu_items`
--

LOCK TABLES `menu_items` WRITE;
/*!40000 ALTER TABLE `menu_items` DISABLE KEYS */;
INSERT INTO `menu_items` VALUES (1,1,'coba','cobaa',1000,5000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/0214705c1a7c3211e8bf756e53130931.jpg',0,'2026-10-05 09:48:43','2026-11-01 13:16:22',NULL),(2,1,'Iced Latte','Perpaduan espresso dan susu segar yang creamy, disajikan dingin.',22000,26000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/ea1994f7f65bbe3d2362229e80a310db.png',0,'2026-10-05 11:14:33','2026-11-01 13:16:22',NULL),(3,1,'Cappuccino','Espresso, susu steam, dan foam lembut dengan rasa klasik.',20000,24000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/443068656ed1e1079998ddae27923af7.png',0,'2026-10-05 11:15:29','2026-11-01 13:16:22',NULL),(4,1,'Americano','Espresso murni dengan air panas, ringan dan tegas.',18000,22000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/251303b831beeb836b5f8e9903fb357d.png',0,'2026-10-05 11:16:22','2026-11-01 13:16:22',NULL),(5,1,'Caramel Latte','Espresso, susu, dan saus karamel manis gurih.',24000,28000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/e83b896bf090f8b43f7323b8955fca42.png',0,'2026-10-05 11:17:17','2026-11-01 13:16:22',NULL),(6,2,'Chocolate','Cokelat pekat yang creamy, cocok untuk pecinta manis.',22000,26000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/0e94d3bb9606c18fd711c1ddccb93aa2.png',0,'2026-10-05 11:19:19','2026-11-01 13:16:22',NULL),(7,2,'Matcha Latte','Teh hijau matcha premium dipadukan dengan susu segar.',24000,28000,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/ce6c329f43a0587defcdaa91d5b4786c.png',0,'2026-10-05 11:20:16','2026-11-01 13:16:22',NULL),(8,3,'Nasi Goreng','Nasi goreng dengan telur, ayam suwir, dan pelengkap spesial.',28000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/000c759ac72b9c5e4bf8961d65cb9b9a.jpg',0,'2026-10-05 11:23:35','2026-10-05 11:23:35',NULL),(9,3,'Chicken Katsu','Ayam katsu crispy disajikan dengan saus dan nasi.',30000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/7d88babc1c718344b0381b8d77248943.jpg',0,'2026-10-05 11:25:30','2026-10-05 11:25:30',NULL),(10,4,'Croissant','Pastry mentega berlapis, renyah di luar dan lembut di dalam.',20000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/28740ddf7211726b20e9265fc3144a8e.jpg',0,'2026-10-05 11:27:32','2026-10-05 11:27:32',NULL),(11,4,'Pain Au Chocolat','Pastry lapis isi cokelat, gurih dan manis berpadu.',18000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/00d248c88d7e44481e2707b172dd2f7f.jpg',0,'2026-10-05 11:29:06','2026-10-05 11:29:06',NULL),(12,4,'Donat','Donat empuk dengan pilihan topping manis.',12000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/70774298527b3c9077411481ed40a98c.jpg',0,'2026-10-05 11:31:25','2026-10-29 18:13:31',NULL),(13,3,'bakso','bakso sapi',25000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/430cf03a8de685b8813589c3cb15628c.jpg',0,'2026-10-29 15:15:36','2026-10-29 15:15:46','2026-10-29 15:15:46'),(14,3,'Bakso','bakso sapi',24000,NULL,'tersedia','https://coffeepastebackend-production.up.railway.app/uploads/menu/9096d6914ce40c79bca54f2c32ae3778.jpg',0,'2026-10-29 18:01:46','2026-10-29 18:02:16','2026-10-29 18:02:16');
/*!40000 ALTER TABLE `menu_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `order_items`
--

DROP TABLE IF EXISTS `order_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `menu_item_id` bigint unsigned DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `price` int unsigned NOT NULL,
  `qty` int unsigned NOT NULL,
  `size` enum('Regular','Large') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ice` enum('Normal','Less','No') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sugar` enum('Normal','Less','No') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_oi_order` (`order_id`),
  KEY `idx_oi_menu` (`menu_item_id`),
  CONSTRAINT `fk_oi_menu` FOREIGN KEY (`menu_item_id`) REFERENCES `menu_items` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_oi_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=129 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `order_items`
--

LOCK TABLES `order_items` WRITE;
/*!40000 ALTER TABLE `order_items` DISABLE KEYS */;
INSERT INTO `order_items` VALUES (1,1,1,'coba','Coffee',1000,1,'Regular','Normal','Less'),(2,2,1,'coba','Coffee',1000,7,'Large','No','Normal'),(3,3,12,'Donat','Pastry',12000,1,NULL,NULL,NULL),(4,3,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Less'),(5,4,6,'Chocolate','Non-Coffee',22000,2,'Regular','No','Normal'),(6,5,3,'Cappuccino','Coffee',20000,1,'Large','Normal','Normal'),(7,6,8,'Nasi Goreng','Makanan',28000,2,NULL,NULL,NULL),(8,6,5,'Caramel Latte','Coffee',24000,1,'Regular','Normal','Normal'),(9,6,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(10,7,6,'Chocolate','Non-Coffee',22000,4,'Regular','Normal','Less'),(11,8,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(12,8,9,'Chicken Katsu','Makanan',30000,2,NULL,NULL,NULL),(13,8,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(14,8,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','Normal'),(15,8,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(16,9,7,'Matcha Latte','Non-Coffee',24000,1,'Large','Normal','Normal'),(17,10,4,'Americano','Coffee',18000,2,'Regular','Normal','Less'),(18,11,3,'Cappuccino','Coffee',20000,1,'Regular','No','Normal'),(19,12,4,'Americano','Coffee',18000,1,'Regular','No','Less'),(20,13,6,'Chocolate','Non-Coffee',22000,1,'Regular','Normal','Less'),(21,14,7,'Matcha Latte','Non-Coffee',24000,1,'Large','Normal','Normal'),(22,15,2,'Iced Latte','Coffee',22000,1,'Regular','No','Normal'),(23,16,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(24,17,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','Normal'),(25,18,4,'Americano','Coffee',18000,1,'Regular','Normal','Less'),(26,19,10,'Croissant','Pastry',20000,2,NULL,NULL,NULL),(27,19,2,'Iced Latte','Coffee',22000,1,'Large','Normal','Less'),(28,19,2,'Iced Latte','Coffee',22000,1,'Regular','No','Normal'),(29,20,2,'Iced Latte','Coffee',22000,2,'Regular','Normal','Normal'),(30,21,5,'Caramel Latte','Coffee',24000,1,'Regular','No','Normal'),(31,22,7,'Matcha Latte','Non-Coffee',24000,2,'Regular','Normal','Normal'),(32,22,6,'Chocolate','Non-Coffee',22000,1,'Regular','Normal','Normal'),(33,22,5,'Caramel Latte','Coffee',24000,2,'Regular','Normal','Normal'),(34,22,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(35,22,3,'Cappuccino','Coffee',20000,2,'Regular','Normal','Normal'),(36,22,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(37,23,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(38,23,5,'Caramel Latte','Coffee',24000,1,'Regular','Normal','Normal'),(39,24,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(40,25,11,'Pain Au Chocolat','Pastry',18000,1,NULL,NULL,NULL),(41,26,12,'Donat','Pastry',12000,1,NULL,NULL,NULL),(42,26,11,'Pain Au Chocolat','Pastry',18000,1,NULL,NULL,NULL),(43,26,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(44,26,9,'Chicken Katsu','Makanan',30000,1,NULL,NULL,NULL),(45,26,8,'Nasi Goreng','Makanan',28000,1,NULL,NULL,NULL),(46,26,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(47,26,6,'Chocolate','Non-Coffee',22000,1,'Regular','Normal','Normal'),(48,26,5,'Caramel Latte','Coffee',24000,1,'Regular','Normal','Normal'),(49,26,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(50,26,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','Normal'),(51,26,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(52,27,7,'Matcha Latte','Non-Coffee',24000,2,'Regular','Normal','Normal'),(53,27,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(54,28,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(55,29,11,'Pain Au Chocolat','Pastry',18000,1,NULL,NULL,NULL),(56,29,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(57,29,9,'Chicken Katsu','Makanan',30000,1,NULL,NULL,NULL),(58,29,3,'Cappuccino','Coffee',20000,2,'Regular','Normal','Normal'),(59,29,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(60,30,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(61,30,6,'Chocolate','Non-Coffee',22000,1,'Regular','Normal','Normal'),(62,30,8,'Nasi Goreng','Makanan',28000,1,NULL,NULL,NULL),(63,31,4,'Americano','Coffee',18000,1,'Large','Less','Normal'),(64,31,2,'Iced Latte','Coffee',22000,1,'Regular','No','Normal'),(65,32,2,'Iced Latte','Coffee',22000,1,'Large','Less','Normal'),(66,33,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(67,34,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(68,35,4,'Americano','Coffee',18000,1,'Regular','Normal','Less'),(69,36,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(70,37,4,'Americano','Coffee',18000,1,'Large','Normal','No'),(71,38,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(72,39,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Less'),(73,40,6,'Chocolate','Non-Coffee',22000,1,'Regular','Less','Normal'),(74,40,6,'Chocolate','Non-Coffee',22000,1,'Large','Normal','Normal'),(75,40,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(76,40,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(77,40,9,'Chicken Katsu','Makanan',30000,1,NULL,NULL,NULL),(78,40,11,'Pain Au Chocolat','Pastry',18000,1,NULL,NULL,NULL),(79,41,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(80,41,6,'Chocolate','Non-Coffee',22000,1,'Regular','Normal','Normal'),(81,41,5,'Caramel Latte','Coffee',24000,1,'Regular','Normal','Normal'),(82,41,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(83,41,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','Normal'),(84,41,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(85,41,12,'Donat','Pastry',12000,1,NULL,NULL,NULL),(86,42,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','Normal'),(87,42,6,'Chocolate','Non-Coffee',22000,1,'Regular','Normal','Normal'),(88,43,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(89,44,10,'Croissant','Pastry',20000,2,NULL,NULL,NULL),(90,45,12,'Donat','Pastry',12000,2,NULL,NULL,NULL),(91,46,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(92,47,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(93,48,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(94,48,2,'Iced Latte','Coffee',22000,1,'Large','Normal','Normal'),(95,49,4,'Americano','Coffee',18000,1,'Regular','Normal','No'),(96,50,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(97,51,4,'Americano','Coffee',18000,1,'Regular','Normal','Normal'),(98,52,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(99,53,12,'Donat','Pastry',12000,1,NULL,NULL,NULL),(100,54,4,'Americano','Coffee',18000,1,'Regular','Normal','Less'),(101,55,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','Normal'),(102,56,12,'Donat','Pastry',12000,1,NULL,NULL,NULL),(103,57,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','No'),(104,58,2,'Iced Latte','Coffee',22000,1,'Large','Normal','Less'),(105,59,9,'Chicken Katsu','Makanan',30000,1,NULL,NULL,NULL),(106,60,12,'Donat','Pastry',12000,2,NULL,NULL,NULL),(107,61,4,'Americano','Coffee',18000,2,'Regular','Normal','Normal'),(108,62,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','No'),(109,63,4,'Americano','Coffee',18000,1,'Regular','Less','Less'),(110,64,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Less'),(111,65,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(112,66,3,'Cappuccino','Coffee',20000,1,'Regular','Normal','Normal'),(113,67,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Normal'),(114,68,12,'Donat','Pastry',12000,1,NULL,NULL,NULL),(115,68,7,'Matcha Latte','Non-Coffee',24000,1,'Regular','Normal','No'),(116,69,2,'Iced Latte','Coffee',22000,1,'Regular','Less','Normal'),(117,70,2,'Iced Latte','Coffee',22000,1,'Large','Normal','Less'),(118,71,8,'Nasi Goreng','Makanan',28000,2,NULL,NULL,NULL),(119,72,11,'Pain Au Chocolat','Pastry',18000,1,NULL,NULL,NULL),(120,72,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(121,73,2,'Iced Latte','Coffee',22000,1,'Large','Normal','Normal'),(122,74,3,'Cappuccino','Coffee',24000,1,'Large','Normal','Less'),(123,74,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(124,75,10,'Croissant','Pastry',20000,1,NULL,NULL,NULL),(125,75,3,'Cappuccino','Coffee',24000,1,'Large','Normal','Less'),(126,76,4,'Americano','Coffee',18000,1,'Regular','Normal','Less'),(127,77,4,'Americano','Coffee',18000,1,'Regular','Less','Less'),(128,78,2,'Iced Latte','Coffee',22000,1,'Regular','Normal','Less');
/*!40000 ALTER TABLE `order_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `order_status_logs`
--

DROP TABLE IF EXISTS `order_status_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_status_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `from_status` enum('menunggu','proses','siap','selesai','dibatalkan') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `to_status` enum('menunggu','proses','siap','selesai','dibatalkan') COLLATE utf8mb4_unicode_ci NOT NULL,
  `changed_by` bigint unsigned DEFAULT NULL,
  `changed_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_osl_order` (`order_id`),
  KEY `fk_osl_user` (`changed_by`),
  CONSTRAINT `fk_osl_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_osl_user` FOREIGN KEY (`changed_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=264 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `order_status_logs`
--

LOCK TABLES `order_status_logs` WRITE;
/*!40000 ALTER TABLE `order_status_logs` DISABLE KEYS */;
INSERT INTO `order_status_logs` VALUES (1,1,NULL,'proses',2,'2026-10-05 09:52:05'),(2,1,'proses','siap',2,'2026-10-05 09:52:35'),(3,1,'siap','selesai',2,'2026-10-05 09:52:42'),(4,2,NULL,'proses',2,'2026-10-05 10:05:28'),(5,3,NULL,'proses',2,'2026-10-05 11:33:30'),(6,4,NULL,'proses',2,'2026-10-05 11:34:09'),(7,5,NULL,'proses',2,'2026-10-05 11:34:26'),(8,6,NULL,'proses',2,'2026-10-05 11:35:07'),(9,2,'proses','siap',2,'2026-10-05 11:35:26'),(10,3,'proses','siap',2,'2026-10-05 11:35:31'),(11,6,'proses','siap',2,'2026-10-05 11:39:39'),(12,3,'siap','selesai',2,'2026-10-05 11:39:48'),(13,7,NULL,'proses',2,'2026-10-05 11:40:40'),(14,8,NULL,'proses',2,'2026-10-05 11:41:18'),(15,8,'proses','siap',2,'2026-10-05 11:41:33'),(16,8,'siap','selesai',2,'2026-10-05 11:41:41'),(17,7,'proses','siap',2,'2026-10-06 08:51:48'),(18,6,'siap','selesai',2,'2026-10-06 08:51:49'),(19,5,'proses','siap',2,'2026-10-06 08:51:51'),(20,5,'siap','selesai',2,'2026-10-06 08:51:52'),(21,4,'proses','siap',2,'2026-10-06 08:51:53'),(22,4,'siap','selesai',2,'2026-10-06 08:51:55'),(23,7,'siap','selesai',2,'2026-10-06 08:51:56'),(24,2,'siap','selesai',2,'2026-10-06 08:51:58'),(25,9,NULL,'proses',2,'2026-10-06 09:27:06'),(26,10,NULL,'proses',2,'2026-10-06 09:27:30'),(27,11,NULL,'proses',2,'2026-10-06 09:28:22'),(28,12,NULL,'proses',2,'2026-10-06 10:43:08'),(29,13,NULL,'proses',2,'2026-10-06 10:44:21'),(30,13,'proses','siap',2,'2026-10-06 10:45:07'),(31,13,'siap','selesai',2,'2026-10-06 10:45:08'),(32,12,'proses','siap',2,'2026-10-06 10:45:09'),(33,12,'siap','selesai',2,'2026-10-06 10:45:10'),(34,14,NULL,'proses',2,'2026-10-06 12:17:11'),(35,15,NULL,'proses',2,'2026-10-06 12:17:28'),(36,15,'proses','siap',2,'2026-10-06 12:17:35'),(37,15,'siap','selesai',2,'2026-10-06 12:17:36'),(38,14,'proses','siap',2,'2026-10-06 12:17:37'),(39,14,'siap','selesai',2,'2026-10-06 12:17:37'),(40,11,'proses','siap',2,'2026-10-06 12:17:38'),(41,11,'siap','selesai',2,'2026-10-06 12:17:38'),(42,10,'proses','siap',2,'2026-10-06 12:17:39'),(43,10,'siap','selesai',2,'2026-10-06 12:17:40'),(44,9,'proses','siap',2,'2026-10-06 12:17:41'),(45,9,'siap','selesai',2,'2026-10-06 12:17:42'),(46,16,NULL,'proses',2,'2026-10-07 08:45:12'),(47,17,NULL,'proses',2,'2026-10-07 08:45:29'),(48,17,'proses','siap',2,'2026-10-07 08:45:36'),(49,17,'siap','selesai',2,'2026-10-07 08:45:37'),(50,16,'proses','siap',2,'2026-10-07 08:45:38'),(51,16,'siap','selesai',2,'2026-10-07 08:45:39'),(52,18,NULL,'proses',2,'2026-10-07 09:52:08'),(53,18,'proses','siap',2,'2026-10-07 09:53:30'),(54,18,'siap','selesai',2,'2026-10-07 09:53:31'),(55,19,NULL,'proses',2,'2026-10-07 10:04:52'),(56,19,'proses','siap',2,'2026-10-07 10:05:09'),(57,19,'siap','selesai',2,'2026-10-07 10:05:12'),(58,20,NULL,'proses',2,'2026-10-07 10:20:07'),(59,21,NULL,'proses',2,'2026-10-07 10:20:25'),(60,21,'proses','siap',2,'2026-10-07 10:20:33'),(61,21,'siap','selesai',2,'2026-10-07 10:20:33'),(62,20,'proses','siap',2,'2026-10-07 10:20:34'),(63,20,'siap','selesai',2,'2026-10-07 10:20:34'),(64,22,NULL,'proses',2,'2026-10-07 10:23:16'),(65,22,'proses','siap',2,'2026-10-07 10:23:27'),(66,22,'siap','selesai',2,'2026-10-07 10:23:28'),(67,23,NULL,'proses',2,'2026-10-07 10:47:38'),(68,23,'proses','siap',2,'2026-10-07 10:47:45'),(69,23,'siap','selesai',2,'2026-10-07 10:47:46'),(70,24,NULL,'proses',2,'2026-10-07 11:15:31'),(71,25,NULL,'proses',2,'2026-10-07 11:58:06'),(72,25,'proses','siap',2,'2026-10-07 11:58:13'),(73,25,'siap','selesai',2,'2026-10-07 11:58:14'),(74,24,'proses','siap',2,'2026-10-07 11:58:15'),(75,24,'siap','selesai',2,'2026-10-07 11:58:15'),(76,26,NULL,'proses',2,'2026-10-08 20:34:51'),(77,26,'proses','siap',2,'2026-10-08 20:35:01'),(78,26,'siap','selesai',2,'2026-10-08 20:35:02'),(79,27,NULL,'proses',2,'2026-10-08 20:37:52'),(80,27,'proses','siap',2,'2026-10-08 20:38:04'),(81,27,'siap','selesai',2,'2026-10-08 20:38:04'),(82,28,NULL,'proses',2,'2026-10-08 20:38:37'),(83,28,'proses','siap',2,'2026-10-08 20:38:45'),(84,28,'siap','selesai',2,'2026-10-08 20:38:46'),(85,29,NULL,'proses',2,'2026-10-09 11:18:14'),(86,29,'proses','siap',2,'2026-10-09 11:18:22'),(87,29,'siap','selesai',2,'2026-10-09 11:18:23'),(88,30,NULL,'proses',2,'2026-10-09 21:20:41'),(89,30,'proses','siap',2,'2026-10-09 21:20:56'),(90,30,'siap','selesai',2,'2026-10-09 21:20:56'),(91,31,NULL,'proses',2,'2026-10-10 11:45:57'),(92,31,'proses','siap',2,'2026-10-10 11:46:16'),(93,31,'siap','selesai',2,'2026-10-10 11:46:17'),(94,32,NULL,'menunggu',3,'2026-10-10 12:48:35'),(95,32,'menunggu','proses',2,'2026-10-10 12:49:05'),(96,32,'proses','siap',2,'2026-10-10 12:49:11'),(97,33,NULL,'menunggu',3,'2026-10-11 06:18:51'),(98,33,'menunggu','proses',2,'2026-10-11 06:19:40'),(99,33,'proses','siap',2,'2026-10-11 06:20:51'),(100,33,'siap','selesai',2,'2026-10-11 06:20:52'),(101,34,NULL,'menunggu',3,'2026-10-11 07:05:17'),(102,34,'menunggu','dibatalkan',3,'2026-10-11 07:05:27'),(103,35,NULL,'menunggu',3,'2026-10-11 07:08:18'),(104,35,'menunggu','proses',2,'2026-10-11 07:08:40'),(105,35,'proses','siap',2,'2026-10-11 07:09:41'),(106,35,'siap','selesai',2,'2026-10-11 07:09:44'),(107,36,NULL,'menunggu',3,'2026-10-11 21:30:25'),(108,36,'menunggu','proses',2,'2026-10-11 21:30:37'),(109,36,'proses','siap',2,'2026-10-11 21:30:38'),(110,36,'siap','selesai',2,'2026-10-11 21:49:26'),(111,37,NULL,'menunggu',3,'2026-10-11 21:50:09'),(112,37,'menunggu','proses',2,'2026-10-11 21:51:03'),(113,37,'proses','siap',2,'2026-10-11 21:51:23'),(114,37,'siap','selesai',2,'2026-10-11 21:51:35'),(115,38,NULL,'proses',2,'2026-10-12 11:42:45'),(116,39,NULL,'proses',2,'2026-10-12 11:43:00'),(117,39,'proses','siap',2,'2026-10-12 11:49:24'),(118,39,'siap','selesai',2,'2026-10-12 11:49:24'),(119,38,'proses','siap',2,'2026-10-12 11:49:25'),(120,38,'siap','selesai',2,'2026-10-12 11:49:25'),(121,40,NULL,'proses',2,'2026-10-12 11:50:13'),(122,40,'proses','siap',2,'2026-10-12 11:50:28'),(123,40,'siap','selesai',2,'2026-10-12 11:50:30'),(124,41,NULL,'proses',2,'2026-10-12 11:58:31'),(125,41,'proses','siap',2,'2026-10-12 11:58:38'),(126,41,'siap','selesai',2,'2026-10-12 11:58:38'),(127,42,NULL,'proses',2,'2026-10-12 12:03:12'),(128,42,'proses','siap',2,'2026-10-12 12:03:18'),(129,42,'siap','selesai',2,'2026-10-12 12:03:18'),(130,43,NULL,'menunggu',4,'2026-10-13 22:47:12'),(131,43,'menunggu','proses',2,'2026-10-13 22:48:27'),(132,43,'proses','siap',2,'2026-10-13 22:48:28'),(133,43,'siap','selesai',2,'2026-10-13 22:48:28'),(134,44,NULL,'menunggu',4,'2026-10-13 22:50:03'),(135,44,'menunggu','proses',2,'2026-10-13 22:50:16'),(136,44,'proses','siap',2,'2026-10-13 22:50:19'),(137,45,NULL,'menunggu',3,'2026-10-14 09:56:44'),(138,46,NULL,'menunggu',4,'2026-10-14 09:57:11'),(139,46,'menunggu','dibatalkan',4,'2026-10-14 09:57:29'),(140,47,NULL,'menunggu',4,'2026-10-14 09:59:20'),(141,45,'menunggu','proses',2,'2026-10-14 10:00:32'),(142,47,'menunggu','proses',2,'2026-10-14 10:00:40'),(143,45,'proses','siap',2,'2026-10-14 10:01:02'),(144,47,'proses','siap',2,'2026-10-14 10:01:10'),(145,45,'siap','selesai',2,'2026-10-14 10:01:19'),(146,47,'siap','selesai',2,'2026-10-14 10:02:05'),(147,32,'siap','selesai',2,'2026-10-14 10:02:12'),(148,44,'siap','selesai',2,'2026-10-14 10:02:15'),(149,48,NULL,'menunggu',3,'2026-10-14 11:42:32'),(150,48,'menunggu','proses',2,'2026-10-14 11:42:54'),(151,49,NULL,'menunggu',3,'2026-10-19 09:57:41'),(152,50,NULL,'menunggu',3,'2026-10-24 20:12:07'),(153,50,'menunggu','proses',2,'2026-10-25 19:06:32'),(154,50,'proses','siap',2,'2026-10-25 19:06:32'),(155,50,'siap','selesai',2,'2026-10-25 19:06:33'),(156,51,NULL,'menunggu',3,'2026-10-25 19:22:57'),(157,51,'menunggu','proses',2,'2026-10-25 19:41:56'),(158,52,NULL,'menunggu',3,'2026-10-25 20:02:38'),(159,52,'menunggu','proses',2,'2026-10-25 20:03:25'),(160,52,'proses','siap',2,'2026-10-25 20:03:31'),(161,52,'siap','selesai',2,'2026-10-25 20:03:35'),(162,53,NULL,'menunggu',3,'2026-10-25 20:05:43'),(163,54,NULL,'menunggu',3,'2026-10-25 20:14:11'),(164,55,NULL,'menunggu',3,'2026-10-25 20:14:39'),(165,55,'menunggu','proses',2,'2026-10-25 20:15:49'),(166,55,'proses','siap',2,'2026-10-25 20:16:21'),(167,55,'siap','selesai',2,'2026-10-25 20:16:23'),(168,54,'menunggu','proses',2,'2026-10-25 20:16:33'),(169,54,'proses','siap',2,'2026-10-25 20:16:36'),(170,53,'menunggu','proses',2,'2026-10-25 20:16:39'),(171,51,'proses','siap',2,'2026-10-25 20:16:43'),(172,56,NULL,'menunggu',5,'2026-10-26 09:27:08'),(173,56,'menunggu','dibatalkan',5,'2026-10-26 09:40:25'),(174,57,NULL,'menunggu',3,'2026-10-26 10:31:39'),(175,57,'menunggu','proses',2,'2026-10-26 10:32:33'),(176,57,'proses','siap',2,'2026-10-26 10:32:40'),(177,57,'siap','selesai',2,'2026-10-26 10:40:34'),(178,54,'siap','selesai',2,'2026-10-26 11:00:03'),(179,53,'proses','siap',2,'2026-10-26 11:00:08'),(180,51,'siap','selesai',2,'2026-10-26 11:00:09'),(181,49,'menunggu','proses',2,'2026-10-26 11:00:11'),(182,49,'proses','siap',2,'2026-10-26 11:00:11'),(183,48,'proses','siap',2,'2026-10-26 11:00:14'),(184,58,NULL,'menunggu',3,'2026-10-27 12:53:29'),(185,58,'menunggu','proses',2,'2026-10-27 12:53:58'),(186,58,'proses','siap',2,'2026-10-27 12:54:23'),(187,58,'siap','selesai',2,'2026-10-27 12:54:48'),(188,59,NULL,'menunggu',3,'2026-10-28 09:24:09'),(189,59,'menunggu','proses',2,'2026-10-28 09:24:26'),(190,59,'proses','siap',2,'2026-10-28 09:24:38'),(191,59,'siap','selesai',2,'2026-10-28 09:24:57'),(192,60,NULL,'menunggu',3,'2026-10-28 12:48:58'),(193,60,'menunggu','proses',2,'2026-10-28 12:49:02'),(194,60,'proses','siap',2,'2026-10-28 12:49:06'),(195,60,'siap','selesai',2,'2026-10-28 12:49:23'),(196,61,NULL,'menunggu',3,'2026-10-28 18:30:24'),(197,61,'menunggu','proses',2,'2026-10-28 18:37:20'),(198,61,'proses','siap',2,'2026-10-28 18:37:21'),(199,61,'siap','selesai',2,'2026-10-29 10:59:49'),(200,53,'siap','selesai',2,'2026-10-29 11:00:18'),(201,49,'siap','selesai',2,'2026-10-29 11:00:26'),(202,48,'siap','selesai',2,'2026-10-29 11:00:37'),(203,62,NULL,'menunggu',3,'2026-10-29 13:01:44'),(204,62,'menunggu','proses',2,'2026-10-29 13:02:15'),(205,62,'proses','siap',2,'2026-10-29 13:02:34'),(206,62,'siap','selesai',2,'2026-10-29 13:02:55'),(207,63,NULL,'menunggu',3,'2026-10-29 13:18:08'),(208,63,'menunggu','proses',2,'2026-10-29 13:18:29'),(209,63,'proses','siap',2,'2026-10-29 13:19:01'),(210,63,'siap','selesai',2,'2026-10-29 13:19:06'),(211,64,NULL,'menunggu',3,'2026-10-29 14:13:17'),(212,64,'menunggu','proses',2,'2026-10-29 14:14:28'),(213,64,'proses','siap',2,'2026-10-29 14:14:49'),(214,64,'siap','selesai',2,'2026-10-29 14:15:13'),(215,65,NULL,'menunggu',3,'2026-10-29 14:15:47'),(216,65,'menunggu','proses',2,'2026-10-29 14:20:51'),(217,65,'proses','siap',2,'2026-10-29 14:20:51'),(218,65,'siap','selesai',2,'2026-10-29 14:20:55'),(219,66,NULL,'menunggu',3,'2026-10-29 14:21:28'),(220,67,NULL,'menunggu',3,'2026-10-29 15:05:36'),(221,67,'menunggu','proses',2,'2026-10-29 15:07:07'),(222,67,'proses','siap',2,'2026-10-29 15:07:43'),(223,67,'siap','selesai',2,'2026-10-29 15:07:45'),(224,66,'menunggu','proses',2,'2026-10-29 15:07:47'),(225,66,'proses','siap',2,'2026-10-29 15:07:47'),(226,66,'siap','selesai',2,'2026-10-29 15:07:49'),(227,68,NULL,'proses',2,'2026-10-29 18:11:16'),(228,68,'proses','siap',2,'2026-10-29 18:12:26'),(229,68,'siap','selesai',2,'2026-10-29 18:12:46'),(230,69,NULL,'menunggu',3,'2026-10-29 18:17:28'),(231,69,'menunggu','proses',2,'2026-10-29 18:21:12'),(232,69,'proses','siap',2,'2026-10-29 18:21:24'),(233,69,'siap','selesai',2,'2026-10-29 18:21:46'),(234,70,NULL,'menunggu',8,'2026-10-29 22:52:32'),(235,70,'menunggu','proses',2,'2026-10-29 22:53:10'),(236,70,'proses','siap',2,'2026-10-29 22:53:14'),(237,70,'siap','selesai',2,'2026-10-29 22:53:16'),(238,71,NULL,'menunggu',10,'2026-10-29 23:11:34'),(239,71,'menunggu','proses',2,'2026-10-29 23:11:57'),(240,71,'proses','siap',2,'2026-10-29 23:11:57'),(241,71,'siap','selesai',2,'2026-10-29 23:12:07'),(242,72,NULL,'menunggu',11,'2026-10-30 09:30:57'),(243,72,'menunggu','dibatalkan',11,'2026-10-30 09:31:02'),(244,73,NULL,'menunggu',3,'2026-11-01 11:52:39'),(245,73,'menunggu','proses',2,'2026-11-01 11:52:52'),(246,73,'proses','siap',2,'2026-11-01 11:53:30'),(247,73,'siap','selesai',2,'2026-11-01 11:53:31'),(248,74,NULL,'menunggu',6,'2026-11-03 10:29:51'),(249,74,'menunggu','proses',2,'2026-11-03 10:30:10'),(250,75,NULL,'menunggu',6,'2026-11-03 10:31:30'),(251,75,'menunggu','proses',2,'2026-11-03 10:31:37'),(252,75,'proses','siap',2,'2026-11-03 10:32:33'),(253,75,'siap','selesai',2,'2026-11-03 10:34:25'),(254,76,NULL,'menunggu',6,'2026-11-03 10:59:43'),(255,76,'menunggu','proses',2,'2026-11-03 10:59:50'),(256,76,'proses','siap',2,'2026-11-03 10:59:55'),(257,77,NULL,'menunggu',6,'2026-11-04 09:47:27'),(258,77,'menunggu','proses',2,'2026-11-04 09:47:36'),(259,77,'proses','siap',2,'2026-11-04 09:47:59'),(260,77,'siap','selesai',2,'2026-11-04 09:48:15'),(261,78,NULL,'menunggu',13,'2026-11-08 08:12:23'),(262,78,'menunggu','proses',2,'2026-11-09 09:08:16'),(263,78,'proses','siap',2,'2026-11-09 09:08:29');
/*!40000 ALTER TABLE `order_status_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `orders`
--

DROP TABLE IF EXISTS `orders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `customer_id` bigint unsigned DEFAULT NULL,
  `customer_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pelanggan',
  `phone` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `source` enum('Mobile','Kasir') COLLATE utf8mb4_unicode_ci NOT NULL,
  `cashier_id` bigint unsigned DEFAULT NULL,
  `cashier_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `order_type` enum('Dine In','Take Away') COLLATE utf8mb4_unicode_ci NOT NULL,
  `table_number` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` enum('menunggu','proses','siap','selesai','dibatalkan') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'menunggu',
  `payment_method` enum('QRIS','Tunai (Cash)') COLLATE utf8mb4_unicode_ci NOT NULL,
  `paid` tinyint(1) NOT NULL DEFAULT '0',
  `paid_at` datetime DEFAULT NULL,
  `cash_received` int unsigned DEFAULT NULL,
  `subtotal` int unsigned NOT NULL,
  `discount` int unsigned NOT NULL DEFAULT '0',
  `total` int unsigned NOT NULL,
  `note` text COLLATE utf8mb4_unicode_ci,
  `cancel_reason` text COLLATE utf8mb4_unicode_ci,
  `cancelled_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_orders_created` (`created_at`),
  KEY `idx_orders_status` (`status`),
  KEY `idx_orders_cashier` (`cashier_id`),
  KEY `idx_orders_customer` (`customer_id`),
  CONSTRAINT `fk_orders_cashier` FOREIGN KEY (`cashier_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_orders_customer` FOREIGN KEY (`customer_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=79 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `orders`
--

LOCK TABLES `orders` WRITE;
/*!40000 ALTER TABLE `orders` DISABLE KEYS */;
INSERT INTO `orders` VALUES (1,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-05 09:52:05',NULL,1000,0,1110,NULL,NULL,NULL,'2026-10-05 09:52:05','2026-10-07 09:43:41'),(2,NULL,'ana',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-05 10:05:28',10000,7000,0,7770,NULL,NULL,NULL,'2026-10-05 10:05:28','2026-10-07 09:43:41'),(3,NULL,'fito',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-05 11:33:30',50000,36000,0,39960,NULL,NULL,NULL,'2026-10-05 11:33:30','2026-10-07 09:43:41'),(4,NULL,'dina',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-05 11:34:09',NULL,44000,0,48840,NULL,NULL,NULL,'2026-10-05 11:34:09','2026-10-07 09:43:41'),(5,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-05 11:34:26',22200,20000,0,22200,NULL,NULL,NULL,'2026-10-05 11:34:26','2026-10-07 09:43:41'),(6,NULL,'maria',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-05 11:35:07',NULL,102000,0,113220,NULL,NULL,NULL,'2026-10-05 11:35:07','2026-10-07 09:43:41'),(7,NULL,'ayu',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-05 11:40:39',100000,88000,0,97680,NULL,NULL,NULL,'2026-10-05 11:40:39','2026-10-07 09:43:41'),(8,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-05 11:41:18',160000,142000,0,157620,NULL,NULL,NULL,'2026-10-05 11:41:18','2026-10-07 09:43:41'),(9,NULL,'tata',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-06 09:27:06',NULL,24000,0,26640,NULL,NULL,NULL,'2026-10-06 09:27:06','2026-10-07 09:43:41'),(10,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-06 09:27:30',NULL,36000,0,39960,NULL,NULL,NULL,'2026-10-06 09:27:30','2026-10-07 09:43:41'),(11,NULL,'ana',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-06 09:28:22',NULL,20000,0,22200,NULL,NULL,NULL,'2026-10-06 09:28:22','2026-10-07 09:43:41'),(12,NULL,'devi',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-06 10:43:08',19980,18000,0,19980,NULL,NULL,NULL,'2026-10-06 10:43:08','2026-10-07 09:43:41'),(13,NULL,'dino',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-06 10:44:21',100000,22000,0,24420,NULL,NULL,NULL,'2026-10-06 10:44:21','2026-10-07 09:43:41'),(14,NULL,'fitri',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-06 12:17:11',50000,24000,0,26640,NULL,NULL,NULL,'2026-10-06 12:17:11','2026-10-07 09:43:41'),(15,NULL,'lala',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-06 12:17:28',NULL,22000,0,24420,NULL,NULL,NULL,'2026-10-06 12:17:28','2026-10-07 09:43:41'),(16,NULL,'yoga',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-07 08:45:12',26640,24000,0,26640,NULL,NULL,NULL,'2026-10-07 08:45:12','2026-10-07 09:43:41'),(17,NULL,'dito',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-07 08:45:28',NULL,20000,0,22200,NULL,NULL,NULL,'2026-10-07 08:45:28','2026-10-07 09:43:41'),(18,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-07 09:52:08',100000,18000,0,19980,NULL,NULL,NULL,'2026-10-07 09:52:08','2026-10-07 09:53:31'),(19,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-07 10:04:52',NULL,84000,0,93240,NULL,NULL,NULL,'2026-10-07 10:04:52','2026-10-07 10:05:12'),(20,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-07 10:20:07',NULL,44000,0,48840,NULL,NULL,NULL,'2026-10-07 10:20:07','2026-10-07 10:20:34'),(21,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-07 10:20:25',30000,24000,0,26640,NULL,NULL,NULL,'2026-10-07 10:20:25','2026-10-07 10:20:33'),(22,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-07 10:23:16',220000,198000,0,219780,NULL,NULL,NULL,'2026-10-07 10:23:16','2026-10-07 10:23:28'),(23,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-07 10:47:38',NULL,46000,0,51060,NULL,NULL,NULL,'2026-10-07 10:47:38','2026-10-07 10:47:46'),(24,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-07 11:15:31',100000,18000,0,19980,NULL,NULL,NULL,'2026-10-07 11:15:31','2026-10-07 11:58:15'),(25,NULL,'Pelanggan',NULL,'Kasir',2,'Maria','Dine In',NULL,'selesai','QRIS',1,'2026-10-07 11:58:06',NULL,18000,0,19980,NULL,NULL,NULL,'2026-10-07 11:58:06','2026-10-07 11:58:14'),(26,NULL,'Pelanggan',NULL,'Kasir',2,'Maria','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-08 20:34:51',264180,238000,0,264180,NULL,NULL,NULL,'2026-10-08 20:34:51','2026-10-08 20:35:02'),(27,NULL,'dinda',NULL,'Kasir',2,'Maria','Dine In',NULL,'selesai','QRIS',1,'2026-10-08 20:37:52',NULL,66000,0,73260,NULL,NULL,NULL,'2026-10-08 20:37:52','2026-10-08 20:38:04'),(28,NULL,'Pelanggan',NULL,'Kasir',2,'Maria','Dine In',NULL,'selesai','QRIS',1,'2026-10-08 20:38:37',NULL,18000,0,19980,NULL,NULL,NULL,'2026-10-08 20:38:37','2026-10-08 20:38:46'),(29,NULL,'maria',NULL,'Kasir',2,'Maria','Dine In',NULL,'selesai','Tunai (Cash)',1,'2026-10-09 11:18:14',200000,130000,0,144300,NULL,NULL,NULL,'2026-10-09 11:18:14','2026-10-09 11:18:23'),(30,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-09 21:20:41',90000,74000,0,82140,NULL,NULL,NULL,'2026-10-09 21:20:41','2026-10-09 21:20:56'),(31,NULL,'dona',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-10 11:45:56',NULL,40000,0,44400,NULL,NULL,NULL,'2026-10-10 11:45:56','2026-10-10 11:46:16'),(32,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-14 10:02:04',NULL,22000,0,24420,NULL,NULL,NULL,'2026-10-10 12:48:35','2026-10-14 10:02:12'),(33,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-11 06:20:08',100000,22000,0,24420,NULL,NULL,NULL,'2026-10-11 06:18:51','2026-10-11 06:20:52'),(34,3,'maria','082132679593','Mobile',NULL,NULL,'Take Away',NULL,'dibatalkan','Tunai (Cash)',0,NULL,NULL,18000,0,19980,NULL,'Dibatalkan oleh pelanggan','2026-10-11 07:05:27','2026-10-11 07:05:17','2026-10-11 07:05:27'),(35,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-11 07:09:19',50000,18000,0,19980,NULL,NULL,NULL,'2026-10-11 07:08:18','2026-10-11 07:09:44'),(36,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-11 21:30:50',24420,22000,0,24420,NULL,NULL,NULL,'2026-10-11 21:30:25','2026-10-11 21:49:26'),(37,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-11 21:50:53',20000,18000,0,19980,NULL,NULL,NULL,'2026-10-11 21:50:09','2026-10-11 21:51:35'),(38,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-12 11:42:45',NULL,20000,0,22200,NULL,NULL,NULL,'2026-10-12 11:42:45','2026-10-12 11:49:25'),(39,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-12 11:43:00',NULL,24000,0,26640,NULL,NULL,NULL,'2026-10-12 11:43:00','2026-10-12 11:49:24'),(40,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-12 11:50:13',NULL,136000,0,150960,NULL,NULL,NULL,'2026-10-12 11:50:13','2026-10-12 11:50:30'),(41,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-12 11:58:31',200000,142000,0,157620,NULL,NULL,NULL,'2026-10-12 11:58:31','2026-10-12 11:58:38'),(42,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-12 12:03:12',60000,46000,0,51060,NULL,NULL,NULL,'2026-10-12 12:03:12','2026-10-12 12:03:18'),(43,4,'ayu','083850906997','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-13 22:48:18',30000,22000,0,24420,NULL,NULL,NULL,'2026-10-13 22:47:12','2026-10-13 22:48:28'),(44,4,'ayu','083850906997','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-13 22:50:28',44400,40000,0,44400,NULL,NULL,NULL,'2026-10-13 22:50:03','2026-10-14 10:02:15'),(45,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-14 10:00:20',NULL,24000,0,26640,NULL,NULL,NULL,'2026-10-14 09:56:44','2026-10-14 10:01:19'),(46,4,'ayuuu','083850906997','Mobile',NULL,NULL,'Take Away',NULL,'dibatalkan','Tunai (Cash)',0,NULL,NULL,20000,0,22200,NULL,'Dibatalkan pelanggan: X','2026-10-14 09:57:29','2026-10-14 09:57:11','2026-10-14 09:57:29'),(47,4,'ayuuu','083850906997','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-14 10:00:40',NULL,18000,0,19980,NULL,NULL,NULL,'2026-10-14 09:59:20','2026-10-14 10:02:05'),(48,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 11:00:35',100000,42000,0,46620,NULL,NULL,NULL,'2026-10-14 11:42:32','2026-10-29 11:00:37'),(49,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-29 11:00:24',NULL,18000,0,19980,NULL,NULL,NULL,'2026-10-19 09:57:41','2026-10-29 11:00:26'),(50,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-25 19:06:26',NULL,22000,0,24420,NULL,NULL,NULL,'2026-10-24 20:12:07','2026-10-25 19:06:33'),(51,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-25 19:41:54',NULL,18000,0,19980,NULL,NULL,NULL,'2026-10-25 19:22:57','2026-10-26 11:00:09'),(52,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-25 20:03:14',100000,20000,0,22200,NULL,NULL,NULL,'2026-10-25 20:02:38','2026-10-25 20:03:35'),(53,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 11:00:16',100000,12000,0,13320,NULL,NULL,NULL,'2026-10-25 20:05:43','2026-10-29 11:00:18'),(54,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-26 11:00:01',50000,18000,0,19980,NULL,NULL,NULL,'2026-10-25 20:14:11','2026-10-26 11:00:03'),(55,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-25 20:15:51',50000,20000,0,22200,NULL,NULL,NULL,'2026-10-25 20:14:39','2026-10-25 20:16:23'),(56,5,'ayu','083850906997','Mobile',NULL,NULL,'Take Away',NULL,'dibatalkan','QRIS',0,NULL,NULL,12000,0,13320,NULL,'Dibatalkan oleh pelanggan','2026-10-26 09:40:25','2026-10-26 09:27:08','2026-10-26 09:40:25'),(57,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-26 10:32:17',100000,20000,0,22200,NULL,NULL,NULL,'2026-10-26 10:31:39','2026-10-26 10:40:34'),(58,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-27 12:53:54',50000,22000,0,24420,NULL,NULL,NULL,'2026-10-27 12:53:29','2026-10-27 12:54:48'),(59,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-28 09:24:56',40000,30000,0,33300,NULL,NULL,NULL,'2026-10-28 09:24:09','2026-10-28 09:24:57'),(60,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-28 12:49:21',50000,24000,0,26640,NULL,NULL,NULL,'2026-10-28 12:48:58','2026-10-28 12:49:22'),(61,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-29 10:59:43',NULL,36000,0,39960,NULL,NULL,NULL,'2026-10-28 18:30:24','2026-10-29 10:59:49'),(62,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 13:01:59',50000,22000,0,24420,NULL,NULL,NULL,'2026-10-29 13:01:44','2026-10-29 13:02:55'),(63,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 13:18:40',19980,18000,0,19980,NULL,NULL,NULL,'2026-10-29 13:18:08','2026-10-29 13:19:06'),(64,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 14:13:54',30000,22000,0,24420,NULL,NULL,NULL,'2026-10-29 14:13:17','2026-10-29 14:15:13'),(65,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-29 14:20:41',NULL,22000,0,24420,NULL,NULL,NULL,'2026-10-29 14:15:47','2026-10-29 14:20:55'),(66,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-29 15:06:10',NULL,20000,0,22200,NULL,NULL,NULL,'2026-10-29 14:21:28','2026-10-29 15:07:49'),(67,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','QRIS',1,'2026-10-29 15:06:33',NULL,22000,0,24420,NULL,NULL,NULL,'2026-10-29 15:05:36','2026-10-29 15:07:45'),(68,NULL,'Pelanggan',NULL,'Kasir',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 18:11:16',40000,36000,0,39960,NULL,NULL,NULL,'2026-10-29 18:11:16','2026-10-29 18:12:46'),(69,3,'maria','082132679593','Mobile',2,'Shinta','Dine In',NULL,'selesai','QRIS',1,'2026-10-29 18:19:57',NULL,22000,0,24420,NULL,NULL,NULL,'2026-10-29 18:17:28','2026-10-29 18:21:46'),(70,8,'ayu','083850906997','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 22:53:08',100000,22000,0,24420,NULL,NULL,NULL,'2026-10-29 22:52:32','2026-10-29 22:53:16'),(71,10,'ayu','083850906997','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-10-29 23:12:04',100000,56000,0,62160,NULL,NULL,NULL,'2026-10-29 23:11:34','2026-10-29 23:12:07'),(72,11,'Wahyu Prasetyo',NULL,'Mobile',NULL,NULL,'Take Away',NULL,'dibatalkan','QRIS',0,NULL,NULL,38000,0,42180,NULL,'Dibatalkan oleh pelanggan','2026-10-30 09:31:02','2026-10-30 09:30:57','2026-10-30 09:31:02'),(73,3,'maria','082132679593','Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-11-01 11:53:11',100000,22000,0,24420,NULL,NULL,NULL,'2026-11-01 11:52:39','2026-11-01 11:53:31'),(74,6,'Ayu Pramudita',NULL,'Mobile',2,'Shinta','Take Away',NULL,'proses','Tunai (Cash)',0,NULL,NULL,44000,0,48840,NULL,NULL,NULL,'2026-11-03 10:29:51','2026-11-03 10:30:10'),(75,6,'Ayu Pramudita',NULL,'Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-11-03 10:34:18',100000,44000,0,48840,NULL,NULL,NULL,'2026-11-03 10:31:30','2026-11-03 10:34:25'),(76,6,'Ayu Pramudita',NULL,'Mobile',2,'Shinta','Take Away',NULL,'siap','QRIS',1,'2026-11-03 11:04:25',NULL,18000,0,19980,NULL,NULL,NULL,'2026-11-03 10:59:43','2026-11-03 11:04:25'),(77,6,'Ayu Pramudita',NULL,'Mobile',2,'Shinta','Take Away',NULL,'selesai','Tunai (Cash)',1,'2026-11-04 09:47:55',50000,18000,0,19980,NULL,NULL,NULL,'2026-11-04 09:47:27','2026-11-04 09:48:15'),(78,13,'Ramzi Farhan Pratama',NULL,'Mobile',2,'Shinta','Take Away',NULL,'siap','Tunai (Cash)',1,'2026-11-09 09:07:52',100000,22000,0,24420,NULL,NULL,NULL,'2026-11-08 08:12:23','2026-11-09 09:08:29');
/*!40000 ALTER TABLE `orders` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `password_reset_otps`
--

DROP TABLE IF EXISTS `password_reset_otps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `password_reset_otps` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `code_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `attempts` tinyint unsigned NOT NULL DEFAULT '0',
  `expires_at` datetime NOT NULL,
  `used_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_pro_user` (`user_id`),
  CONSTRAINT `fk_pro_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `password_reset_otps`
--

LOCK TABLES `password_reset_otps` WRITE;
/*!40000 ALTER TABLE `password_reset_otps` DISABLE KEYS */;
INSERT INTO `password_reset_otps` VALUES (1,3,'149544c4baad0f0e40fdc783f2bb73e3b3f200cab529de7cde0832071b7b1bb0',0,'2026-10-29 10:52:58','2026-10-29 10:43:40','2026-10-29 10:42:58'),(2,3,'aa5e0f1854f1fb80cde70b8df30071d1a7d8a4f0275c4fcf10bb89f9244d04f2',0,'2026-10-29 11:06:03','2026-10-29 10:56:55','2026-10-29 10:56:03'),(6,6,'ba0170b846c9328592f1e106a63c906f4978d6b31bb6f47697fbf6d1f73e122d',0,'2026-11-03 10:04:28','2026-11-03 09:55:28','2026-11-03 09:54:28');
/*!40000 ALTER TABLE `password_reset_otps` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `password_resets`
--

DROP TABLE IF EXISTS `password_resets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `password_resets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `token_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `expires_at` datetime NOT NULL,
  `used_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_pr_user` (`user_id`),
  CONSTRAINT `fk_pr_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `password_resets`
--

LOCK TABLES `password_resets` WRITE;
/*!40000 ALTER TABLE `password_resets` DISABLE KEYS */;
INSERT INTO `password_resets` VALUES (5,2,'4cbb649f6bb96a6d364cacbea4230cd06d4007d6396266628e3faec54c0cecc7','2026-10-09 13:35:53','2026-10-09 12:36:51','2026-10-09 12:35:53'),(14,6,'7bb7b068b43147a457554f8bc953c573f240025c0173a597db24da47764f6b44','2026-10-28 18:35:27',NULL,'2026-10-28 17:35:27'),(15,3,'faa99345595b1fddfaa48dc04ee432716d064b99314359b692f4130f01c3e062','2026-10-28 20:05:33',NULL,'2026-10-28 19:05:33'),(18,2,'eb4ad08a72f1eb17fc8961b0a97a8ff3480733586a804677aaf22446ea18d72f','2026-11-03 11:35:18',NULL,'2026-11-03 10:35:18'),(19,1,'c2cdc1e11a164a1b57a8f363d59130c2ab66e5f1c16e144ce62c8a22981895d2','2026-11-03 11:39:09',NULL,'2026-11-03 10:39:09');
/*!40000 ALTER TABLE `password_resets` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `payments`
--

DROP TABLE IF EXISTS `payments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `payments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL,
  `provider` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `provider_ref` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `amount` int unsigned NOT NULL,
  `status` enum('pending','paid','expired','failed','refunded') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending',
  `qr_string` text COLLATE utf8mb4_unicode_ci,
  `expires_at` datetime DEFAULT NULL,
  `paid_at` datetime DEFAULT NULL,
  `refunded_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_payments_ref` (`provider`,`provider_ref`),
  KEY `idx_payments_order` (`order_id`),
  CONSTRAINT `fk_payments_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=80 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `payments`
--

LOCK TABLES `payments` WRITE;
/*!40000 ALTER TABLE `payments` DISABLE KEYS */;
INSERT INTO `payments` VALUES (1,1,'manual','ORDER-1',1110,'paid',NULL,NULL,'2026-10-05 09:52:05',NULL,'2026-10-05 09:52:05','2026-10-07 09:43:41'),(2,4,'manual','ORDER-4',48840,'paid',NULL,NULL,'2026-10-05 11:34:09',NULL,'2026-10-05 11:34:09','2026-10-07 09:43:41'),(3,6,'manual','ORDER-6',113220,'paid',NULL,NULL,'2026-10-05 11:35:07',NULL,'2026-10-05 11:35:07','2026-10-07 09:43:41'),(4,9,'manual','ORDER-9',26640,'paid',NULL,NULL,'2026-10-06 09:27:06',NULL,'2026-10-06 09:27:06','2026-10-07 09:43:41'),(5,10,'manual','ORDER-10',39960,'paid',NULL,NULL,'2026-10-06 09:27:30',NULL,'2026-10-06 09:27:30','2026-10-07 09:43:41'),(6,11,'manual','ORDER-11',22200,'paid',NULL,NULL,'2026-10-06 09:28:22',NULL,'2026-10-06 09:28:22','2026-10-07 09:43:41'),(7,15,'manual','ORDER-15',24420,'paid',NULL,NULL,'2026-10-06 12:17:28',NULL,'2026-10-06 12:17:28','2026-10-07 09:43:41'),(8,17,'manual','ORDER-17',22200,'paid',NULL,NULL,'2026-10-07 08:45:29',NULL,'2026-10-07 08:45:29','2026-10-07 09:43:41'),(9,19,'manual','ORDER-19',93240,'paid',NULL,NULL,'2026-10-07 10:04:52',NULL,'2026-10-07 10:04:52','2026-10-07 10:04:52'),(10,20,'manual','ORDER-20',48840,'paid',NULL,NULL,'2026-10-07 10:20:07',NULL,'2026-10-07 10:20:07','2026-10-07 10:20:07'),(11,23,'manual','ORDER-23',51060,'paid',NULL,NULL,'2026-10-07 10:47:38',NULL,'2026-10-07 10:47:38','2026-10-07 10:47:38'),(12,25,'manual','ORDER-25',19980,'paid',NULL,NULL,'2026-10-07 11:58:06',NULL,'2026-10-07 11:58:06','2026-10-07 11:58:06'),(13,27,'manual','ORDER-27',73260,'paid',NULL,NULL,'2026-10-08 20:37:52',NULL,'2026-10-08 20:37:52','2026-10-08 20:37:52'),(14,28,'manual','ORDER-28',19980,'paid',NULL,NULL,'2026-10-08 20:38:37',NULL,'2026-10-08 20:38:37','2026-10-08 20:38:37'),(15,31,'manual','ORDER-31',44400,'paid',NULL,NULL,'2026-10-10 11:45:57',NULL,'2026-10-10 11:45:57','2026-10-10 11:45:57'),(16,32,'manual','ORDER-32',24420,'paid',NULL,NULL,'2026-10-14 10:02:04',NULL,'2026-10-10 12:48:35','2026-10-14 10:02:04'),(17,32,'midtrans','CP00032-1791674230',24420,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261010231654KXqYZHsht1ID0703A016304A8D2','2026-10-11 06:46:11',NULL,NULL,'2026-10-11 06:17:11','2026-10-11 06:17:11'),(18,32,'midtrans','CP00032-1791676481',24420,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261010235424zoYwo1YZTiID0703A016304853F','2026-10-11 07:23:42',NULL,NULL,'2026-10-11 06:54:42','2026-10-11 06:54:42'),(19,32,'midtrans','CP00032-1791679444',24420,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261011004347p0908MLSvOID0703A016304F623','2026-10-11 08:13:05',NULL,NULL,'2026-10-11 07:44:05','2026-10-11 07:44:05'),(20,32,'midtrans','CP00032-1791728452',24420,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261011142036adDYg40EkLID0703A0163046652','2026-10-11 21:49:54',NULL,NULL,'2026-10-11 21:20:54','2026-10-11 21:20:54'),(21,32,'midtrans','CP00032-1791730213',24420,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261011144957Go1AHTn3eJID0703A016304928D','2026-10-11 22:19:16',NULL,NULL,'2026-10-11 21:50:16','2026-10-11 21:50:16'),(22,38,'manual','ORDER-38',22200,'paid',NULL,NULL,'2026-10-12 11:42:45',NULL,'2026-10-12 11:42:45','2026-10-12 11:42:45'),(23,39,'manual','ORDER-39',26640,'paid',NULL,NULL,'2026-10-12 11:43:00',NULL,'2026-10-12 11:43:00','2026-10-12 11:43:00'),(24,40,'manual','ORDER-40',150960,'paid',NULL,NULL,'2026-10-12 11:50:13',NULL,'2026-10-12 11:50:13','2026-10-12 11:50:13'),(25,45,'manual','ORDER-45',26640,'paid',NULL,NULL,'2026-10-14 10:00:20',NULL,'2026-10-14 09:56:44','2026-10-14 10:00:20'),(26,47,'manual','ORDER-47',19980,'paid',NULL,NULL,'2026-10-14 10:00:40',NULL,'2026-10-14 09:59:20','2026-10-14 10:00:40'),(30,49,'manual','ORDER-49',19980,'paid',NULL,NULL,'2026-10-29 11:00:24',NULL,'2026-10-19 09:57:41','2026-10-29 11:00:24'),(31,50,'manual','ORDER-50',24420,'paid',NULL,NULL,'2026-10-25 19:06:26',NULL,'2026-10-24 20:12:07','2026-10-25 19:06:26'),(33,49,'midtrans','CP00049-1792930962',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261025122243VTa5o37jU9ID0703A016304F9DB','2026-10-25 19:51:43',NULL,NULL,'2026-10-25 19:22:43','2026-10-25 19:22:43'),(34,51,'manual','ORDER-51',19980,'paid',NULL,NULL,'2026-10-25 19:41:54',NULL,'2026-10-25 19:22:57','2026-10-25 19:41:54'),(35,51,'midtrans','CP00051-1792930977',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261025122257JtJMkw7PEvID0703A0163043ED4','2026-10-25 19:51:57',NULL,NULL,'2026-10-25 19:22:57','2026-10-25 19:22:57'),(37,49,'midtrans','CP00049-1792933349',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261025130229RiKDGcPQCvID0703A0163044C7B','2026-10-25 20:31:29',NULL,NULL,'2026-10-25 20:02:29','2026-10-25 20:02:29'),(38,49,'midtrans','CP00049-1792935999',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261025134640EFwMW18UoBID0703A0163040AB2','2026-10-25 21:15:40',NULL,NULL,'2026-10-25 20:46:40','2026-10-25 20:46:40'),(39,49,'midtrans','CP00049-1792940714',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261025150514tPJfBnxbyRID0703A016304452C','2026-10-25 22:34:15',NULL,NULL,'2026-10-25 22:05:15','2026-10-25 22:05:15'),(40,49,'midtrans','CP00049-1792979972',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261026015932TOK4VBZJaLID0703A01630494F4','2026-10-26 09:28:33',NULL,NULL,'2026-10-26 08:59:33','2026-10-26 08:59:33'),(41,56,'manual','ORDER-56',13320,'failed',NULL,NULL,NULL,NULL,'2026-10-26 09:27:08','2026-10-26 09:40:25'),(42,56,'midtrans','CP00056-1792981628',13320,'failed','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405133205802ID5912coffee paste6007TOMOHON61059543962395028A120261026022708IhTTSGG8GHID0703A016304A22D','2026-10-26 09:56:09',NULL,NULL,'2026-10-26 09:27:09','2026-10-26 09:40:25'),(43,49,'midtrans','CP00049-1792982336',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261026023856NCUtYa7NnhID0703A016304306B','2026-10-26 10:07:57',NULL,NULL,'2026-10-26 09:38:57','2026-10-26 09:38:57'),(44,49,'midtrans','CP00049-1792985504',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261026033145ZtORBHDfQ9ID0703A0163047476','2026-10-26 11:00:45',NULL,NULL,'2026-10-26 10:31:45','2026-10-26 10:31:45'),(45,49,'midtrans','CP00049-1793080365',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261027055246XeQxf9D5ERID0703A016304853A','2026-10-27 13:21:46',NULL,NULL,'2026-10-27 12:52:46','2026-10-27 12:52:46'),(46,49,'midtrans','CP00049-1793103847',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261027122408XVe00StPrtID0703A0163044752','2026-10-27 19:53:08',NULL,NULL,'2026-10-27 19:24:08','2026-10-27 19:24:08'),(47,49,'midtrans','CP00049-1793154201',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261028022322DcEFL4ZxCfID0703A0163048D06','2026-10-28 09:52:22',NULL,NULL,'2026-10-28 09:23:22','2026-10-28 09:23:22'),(48,49,'midtrans','CP00049-1793166505',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A1202610280548269wv06y6w8NID0703A0163043199','2026-10-28 13:17:26',NULL,NULL,'2026-10-28 12:48:26','2026-10-28 12:48:26'),(49,49,'midtrans','CP00049-1793186020',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261028111341mFye1gAVSoID0703A0163049A68','2026-10-28 18:42:41',NULL,NULL,'2026-10-28 18:13:41','2026-10-28 18:13:41'),(50,61,'manual','ORDER-61',39960,'paid',NULL,NULL,'2026-10-29 10:59:43',NULL,'2026-10-28 18:30:24','2026-10-29 10:59:43'),(51,61,'midtrans','CP00061-1793187024',39960,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405399605802ID5912coffee paste6007TOMOHON61059543962395028A120261028113026FDDZ3HufJqID0703A016304305D','2026-10-28 18:59:26',NULL,NULL,'2026-10-28 18:30:26','2026-10-28 18:30:26'),(52,49,'midtrans','CP00049-1793187767',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261028114247H7xctvwaEqID0703A016304E44C','2026-10-28 19:11:48',NULL,NULL,'2026-10-28 18:42:48','2026-10-28 18:42:48'),(53,61,'midtrans','CP00061-1793188766',39960,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405399605802ID5912coffee paste6007TOMOHON61059543962395028A1202610281159271e7ny7J6yGID0703A016304DFC8','2026-10-28 19:28:27',NULL,NULL,'2026-10-28 18:59:27','2026-10-28 18:59:27'),(54,61,'midtrans','CP00061-1793191228',39960,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405399605802ID5912coffee paste6007TOMOHON61059543962395028A120261028124029M3i8BUqCnoID0703A0163048FCD','2026-10-28 20:09:30',NULL,NULL,'2026-10-28 19:40:30','2026-10-28 19:40:30'),(55,49,'midtrans','CP00049-1793191230',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261028124030VEJPR6Pj8XID0703A016304C6E3','2026-10-28 20:09:31',NULL,NULL,'2026-10-28 19:40:31','2026-10-28 19:40:31'),(56,61,'midtrans','CP00061-1793243713',39960,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405399605802ID5912coffee paste6007TOMOHON61059543962395028A120261029031514zt3kle6AaFID0703A0163047F71','2026-10-29 10:44:14',NULL,NULL,'2026-10-29 10:15:14','2026-10-29 10:15:14'),(57,49,'midtrans','CP00049-1793243714',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261029031514GuFoyOvn2nID0703A0163040887','2026-10-29 10:44:15',NULL,NULL,'2026-10-29 10:15:15','2026-10-29 10:15:15'),(58,61,'midtrans','CP00061-1793246234',39960,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405399605802ID5912coffee paste6007TOMOHON61059543962395028A1202610290357141G6dzCk14KID0703A016304B4D7','2026-10-29 11:26:15',NULL,NULL,'2026-10-29 10:57:15','2026-10-29 10:57:15'),(59,49,'midtrans','CP00049-1793246235',19980,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261029035715bwdNBTvFzhID0703A016304CD47','2026-10-29 11:26:15',NULL,NULL,'2026-10-29 10:57:15','2026-10-29 10:57:15'),(62,65,'manual','ORDER-65',24420,'paid',NULL,NULL,'2026-10-29 14:20:41',NULL,'2026-10-29 14:15:47','2026-10-29 14:20:41'),(63,65,'midtrans','CP00065-1793258147',24420,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261029071548AIBF9No4F2ID0703A01630448FE','2026-10-29 14:44:48',NULL,NULL,'2026-10-29 14:15:48','2026-10-29 14:15:48'),(65,66,'manual','ORDER-66',22200,'paid',NULL,NULL,'2026-10-29 15:06:10',NULL,'2026-10-29 14:21:28','2026-10-29 15:06:10'),(66,66,'midtrans','CP00066-1793258488',22200,'paid','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405222005802ID5912coffee paste6007TOMOHON61059543962395028A1202610290721299CHsDlWrmDID0703A0163047D58','2026-10-29 14:50:29','2026-10-29 15:06:10',NULL,'2026-10-29 14:21:29','2026-10-29 15:06:10'),(67,66,'midtrans','CP00066-1793261029',22200,'pending','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405222005802ID5912coffee paste6007TOMOHON61059543962395028A120261029080350MMpadH6VoKID0703A016304D971','2026-10-29 15:32:50',NULL,NULL,'2026-10-29 15:03:50','2026-10-29 15:03:50'),(68,67,'manual','ORDER-67',24420,'paid',NULL,NULL,'2026-10-29 15:06:33',NULL,'2026-10-29 15:05:36','2026-10-29 15:06:33'),(69,67,'midtrans','CP00067-1793261136',24420,'paid','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261029080537gTGu6ERJrGID0703A0163040E2C','2026-10-29 15:34:37','2026-10-29 15:06:33',NULL,'2026-10-29 15:05:37','2026-10-29 15:06:33'),(72,69,'manual','ORDER-69',24420,'paid',NULL,NULL,'2026-10-29 18:19:57',NULL,'2026-10-29 18:17:28','2026-10-29 18:19:57'),(73,69,'midtrans','CP00069-1793272648',24420,'paid','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405244205802ID5912coffee paste6007TOMOHON61059543962395028A120261029111729WWYwkwRnCZID0703A016304EC0C','2026-10-29 18:46:29','2026-10-29 18:19:57',NULL,'2026-10-29 18:17:29','2026-10-29 18:19:57'),(75,72,'manual','ORDER-72',42180,'failed',NULL,NULL,NULL,NULL,'2026-10-30 09:30:57','2026-10-30 09:31:02'),(76,72,'midtrans','CP00072-1793327457',42180,'failed','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405421805802ID5912coffee paste6007TOMOHON61059543962395028A120261030023058ef15jdCb1jID0703A0163042FE0','2026-10-30 09:59:58',NULL,NULL,'2026-10-30 09:30:58','2026-10-30 09:31:02'),(77,76,'manual','ORDER-76',19980,'paid',NULL,NULL,'2026-11-03 11:04:25',NULL,'2026-11-03 10:59:43','2026-11-03 11:04:25'),(78,76,'midtrans','CP00076-1793678383',19980,'paid','00020101021226620014COM.GO-JEK.WWW011993600914344055123100210M4405512310303UKE51440014ID.CO.QRIS.WWW0215AID9762887454750303UKE5204314653033605405199805802ID5912coffee paste6007TOMOHON61059543962395028A120261103035944JRCNidvhWiID0703A0163046652','2026-11-03 11:28:44','2026-11-03 11:04:25',NULL,'2026-11-03 10:59:44','2026-11-03 11:04:25');
/*!40000 ALTER TABLE `payments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `reviews`
--

DROP TABLE IF EXISTS `reviews`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `reviews` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `order_id` bigint unsigned DEFAULT NULL,
  `menu_item_id` bigint unsigned DEFAULT NULL,
  `rating` tinyint unsigned NOT NULL,
  `comment` text COLLATE utf8mb4_general_ci NOT NULL,
  `is_published` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_review_order_item` (`order_id`,`menu_item_id`),
  KEY `idx_reviews_published` (`is_published`,`created_at`),
  KEY `idx_reviews_user` (`user_id`),
  KEY `idx_reviews_menu_item` (`menu_item_id`,`is_published`),
  CONSTRAINT `fk_reviews_menu_item` FOREIGN KEY (`menu_item_id`) REFERENCES `menu_items` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_reviews_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_reviews_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `reviews`
--

LOCK TABLES `reviews` WRITE;
/*!40000 ALTER TABLE `reviews` DISABLE KEYS */;
INSERT INTO `reviews` VALUES (1,3,33,2,5,'rasanya enakkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk',1,'2026-10-11 06:21:45','2026-10-11 06:21:45'),(2,4,47,4,5,'xxxx',1,'2026-10-14 10:13:12','2026-10-14 10:13:12'),(3,3,45,12,4,'donatnya bundar',1,'2026-10-14 10:14:00','2026-10-14 10:14:00'),(4,3,37,4,1,'rasanya pahitt banget, tapi enak sih',1,'2026-10-14 10:14:51','2026-10-14 10:14:51'),(5,3,69,2,5,'enak sekali iced lattenya, tempatnya juga nyaman',1,'2026-10-29 18:22:55','2026-10-29 18:22:55'),(6,8,70,2,5,'enak puolllllllllllllllllllllll',1,'2026-10-29 22:54:01','2026-10-29 22:54:01');
/*!40000 ALTER TABLE `reviews` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `stock_shortages`
--

DROP TABLE IF EXISTS `stock_shortages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `stock_shortages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(80) COLLATE utf8mb4_unicode_ci NOT NULL,
  `category` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `note` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` enum('habis','dibeli') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'habis',
  `reported_by` bigint unsigned DEFAULT NULL,
  `reported_by_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `reported_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `purchased_by` bigint unsigned DEFAULT NULL,
  `purchased_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_stock_status` (`status`,`reported_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `stock_shortages`
--

LOCK TABLES `stock_shortages` WRITE;
/*!40000 ALTER TABLE `stock_shortages` DISABLE KEYS */;
INSERT INTO `stock_shortages` VALUES (1,'Cup Plastik','Kemasan','sisa 1 lusin, perlu segera','dibeli',2,'Shinta','2026-11-01 18:48:25',1,'2026-11-01 18:51:22'),(2,'Sirup','Bahan Baku','sisa 1 botol, perlu segera','dibeli',2,'Shinta','2026-11-01 18:48:47',1,'2026-11-01 18:50:50'),(3,'Gula','Bahan Baku','','dibeli',2,'Shinta','2026-11-01 18:48:55',1,'2026-11-01 18:50:41');
/*!40000 ALTER TABLE `stock_shortages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role` enum('admin','kasir','customer') COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(150) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `google_sub` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `token_version` int unsigned NOT NULL DEFAULT '0',
  `phone` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `address` text COLLATE utf8mb4_unicode_ci,
  `gender` enum('Laki-laki','Perempuan') COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `birth_place` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `birth_date` date DEFAULT NULL,
  `photo_url` text COLLATE utf8mb4_unicode_ci,
  `is_active` tinyint(1) NOT NULL DEFAULT '1',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_users_email` (`email`),
  UNIQUE KEY `uq_users_google_sub` (`google_sub`),
  KEY `idx_users_role` (`role`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'admin','Admin','shintamaria9265@gmail.com','$2a$10$q3GNva3Rd94pK6B98U7AoOAHl9LOzuiXoRhLrXAh3/NCxo.ZvoJhe',NULL,0,'0812-3456-7890',NULL,NULL,NULL,NULL,NULL,1,'2026-10-03 19:07:14','2026-10-03 22:06:29',NULL),(2,'kasir','Shinta','shintamaria293@gmail.com','$2a$10$acBBZawadSWUNQuXQFQDjuD3DftcI3ypRRK36Gd7A38PnjNfS3AIO',NULL,2,'0812 3456 7890','Jl. Melati No. 10, Jember','Perempuan',NULL,NULL,NULL,1,'2026-10-03 19:07:14','2026-10-09 12:40:19',NULL),(3,'customer','maria','shimar2901@gmail.com','$2a$10$mP9DMoS7Kn0Yi7mgV1MIBurF7SbUz.tOjIi3qApkovHIec5.CZ0qm','107902368443164247802',2,'082132679593',NULL,'Perempuan',NULL,NULL,NULL,1,'2026-10-10 12:46:37','2026-10-29 10:56:55',NULL),(4,'customer','ayuuu','ayupdita789@gmail.com','$2a$10$2i2qWokCH8VB5mMlbBM/QuRmIctLIJQ8IkPIoofF4u.fZaNH.y7wa','117497151793313694800',1,'083850906997',NULL,NULL,NULL,NULL,NULL,1,'2026-10-13 22:46:12','2026-10-29 23:14:16',NULL),(5,'customer','ayu','e41230760@student.polije.ac.id','$2a$10$y9fsC5oAxXEu2jFsaDdvWOJzEupUvLFr4KCf4Aw.OOXI1mjlYJ2FG',NULL,0,'083850906997',NULL,NULL,NULL,NULL,NULL,1,'2026-10-20 19:19:27','2026-10-20 19:19:27',NULL),(6,'customer','Ayu Pramudita','ayupramudita63@gmail.com','$2a$10$6C0vMNjQ.9b/bFukeWMDSuIsM4muLTnUXjdVEvsV3mSm22HxE5b56','105458700370697801392',1,NULL,NULL,NULL,NULL,'2004-01-01',NULL,1,'2026-10-27 10:11:50','2026-11-03 10:03:06',NULL),(8,'customer','ayu','apc129d6x0530@student.devacademy.id','$2a$10$mC1jeVFdY0ZhNjbWescS0unOH3t4jd6qdht5KY1o7RrHl1LLvszdy',NULL,0,'083850906997',NULL,NULL,NULL,NULL,NULL,1,'2026-10-29 13:16:05','2026-10-29 13:16:05',NULL),(9,'customer','ayu','ayulramidita63@gmail.com','$2a$10$1qa561hy0Y4keg4aEp.RPeEKuDwAWBmrulI8qjsHlSY4I3Aey2u4q',NULL,0,'083850906997',NULL,NULL,NULL,NULL,NULL,1,'2026-10-29 23:08:44','2026-10-29 23:08:44',NULL),(10,'customer','ayu','apdita63@gmail.com','$2a$10$gCE0hqswmb6TQL5tzOXnkeZFGxLfmD2CyGEMGHHeOT1tQ.q3sh8gy',NULL,0,'083850906997',NULL,NULL,NULL,NULL,NULL,1,'2026-10-29 23:11:06','2026-10-29 23:11:06',NULL),(11,'customer','Wahyu Prasetyo','hanyawahyusaja@gmail.com','$2a$10$W19sO0PLoDr9Us9CMk234.T3j8qY33506r7nt0H5X.NmpSr6jhNrm','111180250419040529314',0,NULL,NULL,NULL,NULL,NULL,NULL,1,'2026-10-30 09:30:46','2026-10-30 09:30:46',NULL),(12,'customer','Ramzi','ramzifarhan156@gmail.com','$2a$10$efSVLVPc1y7FCDwQrSA6i.ceT4CY9VvroWBkoYwoChgnlLC75RDhy',NULL,0,'0895706567010',NULL,NULL,NULL,NULL,NULL,1,'2026-11-03 11:23:23','2026-11-03 11:23:23',NULL),(13,'customer','Ramzi Farhan Pratama','gamingsniper094@gmail.com','$2a$10$CTHQeAL/qUaflbHo.0DyBOly4hv2wSLVKs5ia9/QHS.8mgJai9fc2','116485782320924124397',0,NULL,NULL,NULL,NULL,NULL,NULL,1,'2026-11-08 08:11:35','2026-11-08 08:11:35',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Temporary view structure for view `v_customers`
--

DROP TABLE IF EXISTS `v_customers`;
/*!50001 DROP VIEW IF EXISTS `v_customers`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `v_customers` AS SELECT 
 1 AS `id`,
 1 AS `name`,
 1 AS `email`,
 1 AS `phone`,
 1 AS `joined_at`,
 1 AS `total_orders`,
 1 AS `total_spend`*/;
SET character_set_client = @saved_cs_client;

--
-- Temporary view structure for view `v_menu_best_sellers`
--

DROP TABLE IF EXISTS `v_menu_best_sellers`;
/*!50001 DROP VIEW IF EXISTS `v_menu_best_sellers`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `v_menu_best_sellers` AS SELECT 
 1 AS `id`,
 1 AS `category_id`,
 1 AS `name`,
 1 AS `description`,
 1 AS `price`,
 1 AS `status`,
 1 AS `image_url`,
 1 AS `total_sold`*/;
SET character_set_client = @saved_cs_client;

--
-- Temporary view structure for view `v_orders`
--

DROP TABLE IF EXISTS `v_orders`;
/*!50001 DROP VIEW IF EXISTS `v_orders`*/;
SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `v_orders` AS SELECT 
 1 AS `id`,
 1 AS `customer_id`,
 1 AS `customer_name`,
 1 AS `phone`,
 1 AS `source`,
 1 AS `cashier_id`,
 1 AS `order_type`,
 1 AS `table_number`,
 1 AS `status`,
 1 AS `payment_method`,
 1 AS `paid`,
 1 AS `paid_at`,
 1 AS `cash_received`,
 1 AS `subtotal`,
 1 AS `discount`,
 1 AS `total`,
 1 AS `note`,
 1 AS `cancel_reason`,
 1 AS `cancelled_at`,
 1 AS `created_at`,
 1 AS `updated_at`,
 1 AS `code`,
 1 AS `cashier_name`*/;
SET character_set_client = @saved_cs_client;

--
-- Final view structure for view `v_customers`
--

/*!50001 DROP VIEW IF EXISTS `v_customers`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_0900_ai_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`%` SQL SECURITY DEFINER */
/*!50001 VIEW `v_customers` AS select `u`.`id` AS `id`,`u`.`name` AS `name`,`u`.`email` AS `email`,`u`.`phone` AS `phone`,`u`.`created_at` AS `joined_at`,count(`o`.`id`) AS `total_orders`,coalesce(sum(`o`.`total`),0) AS `total_spend` from (`users` `u` left join `orders` `o` on(((`o`.`customer_id` = `u`.`id`) and (`o`.`status` = 'selesai')))) where ((`u`.`role` = 'customer') and (`u`.`deleted_at` is null)) group by `u`.`id`,`u`.`name`,`u`.`email`,`u`.`phone`,`u`.`created_at` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Final view structure for view `v_menu_best_sellers`
--

/*!50001 DROP VIEW IF EXISTS `v_menu_best_sellers`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_0900_ai_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`%` SQL SECURITY DEFINER */
/*!50001 VIEW `v_menu_best_sellers` AS select `m`.`id` AS `id`,`m`.`category_id` AS `category_id`,`m`.`name` AS `name`,`m`.`description` AS `description`,`m`.`price` AS `price`,`m`.`status` AS `status`,`m`.`image_url` AS `image_url`,coalesce(sum((case when (`o`.`id` is not null) then `oi`.`qty` else 0 end)),0) AS `total_sold` from ((`menu_items` `m` left join `order_items` `oi` on((`oi`.`menu_item_id` = `m`.`id`))) left join `orders` `o` on(((`o`.`id` = `oi`.`order_id`) and (`o`.`status` = 'selesai')))) where (`m`.`deleted_at` is null) group by `m`.`id`,`m`.`category_id`,`m`.`name`,`m`.`description`,`m`.`price`,`m`.`status`,`m`.`image_url` */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Final view structure for view `v_orders`
--

/*!50001 DROP VIEW IF EXISTS `v_orders`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = utf8mb4 */;
/*!50001 SET character_set_results     = utf8mb4 */;
/*!50001 SET collation_connection      = utf8mb4_0900_ai_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`%` SQL SECURITY DEFINER */
/*!50001 VIEW `v_orders` AS select `o`.`id` AS `id`,`o`.`customer_id` AS `customer_id`,`o`.`customer_name` AS `customer_name`,`o`.`phone` AS `phone`,`o`.`source` AS `source`,`o`.`cashier_id` AS `cashier_id`,`o`.`order_type` AS `order_type`,`o`.`table_number` AS `table_number`,`o`.`status` AS `status`,`o`.`payment_method` AS `payment_method`,`o`.`paid` AS `paid`,`o`.`paid_at` AS `paid_at`,`o`.`cash_received` AS `cash_received`,`o`.`subtotal` AS `subtotal`,`o`.`discount` AS `discount`,`o`.`total` AS `total`,`o`.`note` AS `note`,`o`.`cancel_reason` AS `cancel_reason`,`o`.`cancelled_at` AS `cancelled_at`,`o`.`created_at` AS `created_at`,`o`.`updated_at` AS `updated_at`,concat('#CP',lpad(`o`.`id`,5,'0')) AS `code`,coalesce(`c`.`name`,`o`.`cashier_name`) AS `cashier_name` from (`orders` `o` left join `users` `c` on((`c`.`id` = `o`.`cashier_id`))) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-11-10  9:11:41
