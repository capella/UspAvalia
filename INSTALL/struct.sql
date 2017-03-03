-- phpMyAdmin SQL Dump
-- version 4.0.10.14
-- http://www.phpmyadmin.net
--
-- Host: localhost:3306
-- Generation Time: Mar 03, 2017 at 08:53 AM
-- Server version: 5.5.37-35.1
-- PHP Version: 5.4.31

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;

--
-- Database: `capeocom_uspavalia`
--

DELIMITER $$
--
-- Procedures
--
$$

--
-- Functions
--
$$

DELIMITER ;

-- --------------------------------------------------------

--
-- Table structure for table `aulaprofessor`
--

CREATE TABLE IF NOT EXISTS `aulaprofessor` (
  `id` int(100) NOT NULL AUTO_INCREMENT,
  `idaula` int(100) NOT NULL,
  `idprofessor` int(100) NOT NULL,
  `uso` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `time` int(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idaula` (`idaula`,`idprofessor`),
  KEY `id` (`id`),
  KEY `id_2` (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=29069 ;

-- --------------------------------------------------------

--
-- Table structure for table `cometario`
--

CREATE TABLE IF NOT EXISTS `cometario` (
  `id` int(100) NOT NULL AUTO_INCREMENT,
  `iduso` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `comantario` varchar(800) COLLATE utf8_unicode_ci NOT NULL,
  `aulaprofessorid` int(100) NOT NULL,
  `time` int(100) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `id` (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=403 ;

-- --------------------------------------------------------

--
-- Table structure for table `disciplinas`
--

CREATE TABLE IF NOT EXISTS `disciplinas` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `nome` varchar(500) COLLATE utf8_unicode_ci NOT NULL,
  `codigo` varchar(40) COLLATE utf8_unicode_ci NOT NULL,
  `idunidade` int(11) NOT NULL,
  `roubo` int(10) NOT NULL DEFAULT '0',
  `uso` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `time` int(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `codigo` (`codigo`,`idunidade`),
  KEY `id` (`id`),
  FULLTEXT KEY `nome` (`nome`,`codigo`)
) ENGINE=MyISAM  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=11017 ;

-- --------------------------------------------------------

--
-- Stand-in structure for view `ListaMedias`
--
CREATE TABLE IF NOT EXISTS `ListaMedias` (
`APid` int(255)
,`AVG(nota)` decimal(14,4)
,`COUNT(*)` bigint(21)
);
-- --------------------------------------------------------

--
-- Stand-in structure for view `Melhores`
--
CREATE TABLE IF NOT EXISTS `Melhores` (
`media` decimal(15,4)
,`votos` bigint(21)
,`materia` varchar(500)
,`unidade` varchar(400)
,`codigo` varchar(40)
,`professor` varchar(300)
,`id` int(100)
);
-- --------------------------------------------------------

--
-- Table structure for table `professores`
--

CREATE TABLE IF NOT EXISTS `professores` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `nome` varchar(300) COLLATE utf8_unicode_ci NOT NULL,
  `idunidade` int(10) NOT NULL,
  `uso` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `time` int(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `nome` (`nome`),
  KEY `id` (`id`),
  FULLTEXT KEY `nome_2` (`nome`)
) ENGINE=MyISAM  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=6509 ;

-- --------------------------------------------------------

--
-- Table structure for table `unidades`
--

CREATE TABLE IF NOT EXISTS `unidades` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `NOME` varchar(400) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  KEY `id` (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=100 ;

-- --------------------------------------------------------

--
-- Table structure for table `votos`
--

CREATE TABLE IF NOT EXISTS `votos` (
  `id` int(100) NOT NULL AUTO_INCREMENT,
  `APid` int(255) NOT NULL,
  `iduso` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `time` int(255) NOT NULL,
  `nota` int(5) NOT NULL,
  `tipo` int(10) NOT NULL DEFAULT '1' COMMENT '1-geral;2-didatica;3-empenho;4-relacaoaluno;5dificuldade',
  PRIMARY KEY (`id`),
  UNIQUE KEY `APid` (`APid`,`iduso`,`tipo`),
  KEY `id` (`id`)
) ENGINE=MyISAM  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=32557 ;

-- --------------------------------------------------------

--
-- Table structure for table `votoscomentario`
--

CREATE TABLE IF NOT EXISTS `votoscomentario` (
  `id` int(100) NOT NULL AUTO_INCREMENT,
  `idcomentario` int(100) NOT NULL,
  `time` int(100) NOT NULL,
  `voto` int(100) NOT NULL COMMENT '-1 ou 1',
  `iduso` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idcomentario` (`idcomentario`,`iduso`),
  KEY `id` (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci AUTO_INCREMENT=503 ;

-- --------------------------------------------------------

--
-- Structure for view `ListaMedias`
--
DROP TABLE IF EXISTS `ListaMedias`;

CREATE ALGORITHM=UNDEFINED DEFINER=`capeocom_uspava`@`201.93.%.%` SQL SECURITY DEFINER VIEW `ListaMedias` AS select `v`.`APid` AS `APid`,avg(`v`.`nota`) AS `AVG(nota)`,count(0) AS `COUNT(*)` from `votos` `v` where (`v`.`tipo` <> 5) group by `v`.`APid`;

-- --------------------------------------------------------

--
-- Structure for view `Melhores`
--
DROP TABLE IF EXISTS `Melhores`;

CREATE ALGORITHM=UNDEFINED DEFINER=`capeocom_uspava`@`201.93.%.%` SQL SECURITY DEFINER VIEW `Melhores` AS select (`ListaMedias`.`AVG(nota)` * 2) AS `media`,`ListaMedias`.`COUNT(*)` AS `votos`,`disciplinas`.`nome` AS `materia`,`unidades`.`NOME` AS `unidade`,`disciplinas`.`codigo` AS `codigo`,`professores`.`nome` AS `professor`,`aulaprofessor`.`id` AS `id` from ((((`ListaMedias` join `aulaprofessor` on((`ListaMedias`.`APid` = `aulaprofessor`.`id`))) join `disciplinas` on((`aulaprofessor`.`idaula` = `disciplinas`.`id`))) join `unidades` on((`disciplinas`.`idunidade` = `unidades`.`id`))) join `professores` on((`aulaprofessor`.`idprofessor` = `professores`.`id`))) where (`ListaMedias`.`COUNT(*)` >= 15) order by `ListaMedias`.`AVG(nota)` desc,`ListaMedias`.`COUNT(*)` desc;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
