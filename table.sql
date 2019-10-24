 CREATE TABLE `two_color_ball` (
  `no` int(10) unsigned NOT NULL COMMENT "开奖期号",
  `red1` tinyint(2) unsigned NOT NULL,
  `red2` tinyint(2) unsigned NOT NULL,
  `red3` tinyint(2) unsigned NOT NULL,
  `red4` tinyint(2) unsigned NOT NULL,
  `red5` tinyint(2) unsigned NOT NULL,
  `red6` tinyint(2) unsigned NOT NULL,
  `Blue1` tinyint(2) unsigned NOT NULL,
  `create_ts` int(10) unsigned NOT NULL,
  `date` char(10) NOT NULL COMMENT "开奖时间",
  PRIMARY KEY (`no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 
