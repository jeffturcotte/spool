<?php
$db = new PDO('pgsql:dbname=postgres;host=db', 'postgres');

$db->query('CREATE TABLE IF NOT EXISTS test (id integer);');
$db->query('START TRANSACTION;');
$db->query('INSERT INTO test(id) VALUES (' . rand(1, 50) . ');');
$db->query('COMMIT;');

$data = $db->query("SELECT count(*) FROM test;")->fetchAll();
?>


<h1>Your number is: <?= $data[0][0] ?></h1>
