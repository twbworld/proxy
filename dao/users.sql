CREATE DATABASE `trojan`;
USE `trojan`;
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `username` varchar(64) NOT NULL,
    `password` char(56) NOT NULL,
    `passwordShow` varchar(255) NOT NULL,
    `quota` bigint NOT NULL DEFAULT '0',
    `download` bigint unsigned NOT NULL DEFAULT '0',
    `upload` bigint unsigned NOT NULL DEFAULT '0',
    `useDays` int DEFAULT '0',
    `expiryDate` char(10) DEFAULT '',
    PRIMARY KEY (`id`),
    KEY `password` (`password`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';
INSERT INTO `users` (`username`, `password`, `passwordShow`, `quota`) VALUES('test', '90a3ed9e32b2aaf4c61c410eb925426119e1a9dc53d4286ade99a809', 'OTBhM2VkOWUzMmIyYWFmNGM2MWM0MTBlYjkyNTQyNjExOWUxYTlkYzUzZDQyODZhZGU5OWE4MDk=', '-1');
