-- for first time run the following command  from terminal
-- psql postgres

-- we are createing a user called mbadmin , with password 'kalra' and his role is to allow create db
drop role mbadmin;
create role mbadmin with login password 'kalra' superuser createdb;

-- next we go to create database 
create database message_board;
\c message_board;

--next we create table in database
drop type user_status;
drop type message_status;
create type user_status as enum ('active', 'blocked', 'admin_blocked');
create type message_status as enum ('active', 'deleted', 'expired', 'admin_blocked');
drop table messages;
drop table users;

create table users (
		user_id bigserial not null,
		firstname varchar(100),
		lastname  varchar(100),
		email     varchar(100),
		password  varchar(100),
		status    user_status,
		reason    varchar(1000),
		location  varchar(100),
		phone     varchar(100),
		preferences varchar(100),
		one_time_token varchar(100),
		created_on_date date,
		created_on_time time,
		primary key(email)
		);

create table messages (
		message_id bigserial not null,
		created_by_user varchar(100) references users(email),
		message varchar(1000), 
		title varchar(100),
		status message_status,
		created_on_date date,
		created_on_time time,
		primary key(message_id)
		);

insert into users (firstname, lastname, email, password) 
	values 
	('admin', 'sandeep kalra', 'sandeep.kalra@gmail.com', md5('nonce256')
	);


