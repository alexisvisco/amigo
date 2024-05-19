--
-- PostgreSQL database dump
--

-- Dumped from database version 16.0
-- Dumped by pg_dump version 16.2

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: migrations_with_change; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA migrations_with_change;


SET default_table_access_method = heap;

--
-- Name: mig_schema_versions; Type: TABLE; Schema: migrations_with_change; Owner: -
--

CREATE TABLE migrations_with_change.mig_schema_versions (
    id text NOT NULL
);


--
-- Name: mig_schema_versions mig_schema_versions_pkey; Type: CONSTRAINT; Schema: migrations_with_change; Owner: -
--

ALTER TABLE ONLY migrations_with_change.mig_schema_versions
    ADD CONSTRAINT mig_schema_versions_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

