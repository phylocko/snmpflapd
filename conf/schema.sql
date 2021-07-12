DROP TABLE IF EXISTS `ports`;
CREATE TABLE `ports`
(
    `id`            int(11) NOT NULL AUTO_INCREMENT,
    `sid`           char(50),
    `timeTicks`     bigint(12),
    `time`          datetime     DEFAULT NULL,
    `ipaddress`     varchar(255) DEFAULT NULL,
    `hostname`      varchar(255) DEFAULT NULL,
    `ifIndex`       int(8)  NOT NULL,
    `ifName`        varchar(255) DEFAULT NULL,
    `ifAlias`       varchar(255) DEFAULT NULL,
    `ifAdminStatus` varchar(255) DEFAULT NULL,
    `ifOperStatus`  varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `time` (`time`)
);

DROP TABLE IF EXISTS `cache_hostname`;
CREATE TABLE `cache_hostname`
(
    `id`        int(11)  NOT NULL AUTO_INCREMENT,
    `time`      datetime NOT NULL default now(),
    `ipaddress` varchar(255)      DEFAULT NULL UNIQUE,
    `hostname`  varchar(255)      DEFAULT NULL,
    PRIMARY KEY (`id`)
);

DROP TABLE IF EXISTS `cache_ifname`;
CREATE TABLE `cache_ifname`
(
    `id`        int(11)      NOT NULL AUTO_INCREMENT,
    `time`      datetime     NOT NULL default now(),
    `ipaddress` varchar(255) NOT NULL,
    `ifIndex`   int(8)       NOT NULL,
    `ifName`    varchar(50)  NOT NULL,
    PRIMARY KEY (`id`)
);

DROP TABLE IF EXISTS `cache_ifalias`;
CREATE TABLE `cache_ifalias`
(
    `id`        int(11)      NOT NULL AUTO_INCREMENT,
    `time`      datetime     NOT NULL default now(),
    `ipaddress` varchar(255) NOT NULL,
    `ifIndex`   int(8)       NOT NULL,
    `ifAlias`   varchar(50)           DEFAULT NULL,
    PRIMARY KEY (`id`)
);
CREATE INDEX idx_sid USING btree ON ports (sid);
CREATE INDEX idx_time USING btree ON ports (time);
