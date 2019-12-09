<?php
//todo add prepared function
//todo edit config loader
//$user $passwod

/**
 * DBConnection Class
 * load $user $password $salt from config.php (edit function later)
 *
 * @access public
 * @author douro
 * @category Database
 * @package kakecoder
 */

Class DBC {
    private $dbh;
    private $dsn;
    private $config;      

    function __construct()
    {
        try{
            $this->config = include_once("config.php");      
            $this->dsn = "mysql:dbname=kakecoder;host=localhost;charset=utf8";
            $this->dbh = new PDO($this->dsn, $this->config["user"], $this->config["pass"]);
            $this->dbh->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        }catch(Exception $e){
            echo "DB INIT ERROR 1";
            exit();
        }
    }

/* 
 * @access public
 * @param string $sql
 * @param array $data
 * @return array $rec
 * @todo catch exceptions
 *
 **/
    function prepare_execute_oneline($sql, $data){
        $stmt = $this->dbh->prepare($sql);
        $stmt->execute($data); 
        try{
            $rec = $stmt->fetch(PDO::FETCH_ASSOC); 
        }catch(Exception $e){
            return true;
        }
        return $rec;
    }
    function simple_exec_obj($sql) {
        $stmt = $this->dbh->query($sql);
        return $stmt;
    }
    function sha256hash($input){
        return hash("sha256",$input.$this->config["salt"]);
    }

    function __destruct()
    {
        $this->dsn = null;
    }
}

?>
