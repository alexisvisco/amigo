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
-- Name: migrations_with_classic; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA migrations_with_classic;


SET default_table_access_method = heap;

--
-- Name: mig_schema_versions; Type: TABLE; Schema: migrations_with_classic; Owner: -
--

CREATE TABLE migrations_with_classic.mig_schema_versions (
    version text NOT NULL
);


--
-- Name: mig_schema_versions mig_schema_versions_pkey; Type: CONSTRAINT; Schema: migrations_with_classic; Owner: -
--

ALTER TABLE ONLY migrations_with_classic.mig_schema_versions
    ADD CONSTRAINT mig_schema_versions_pkey PRIMARY KEY (version);


--
-- PostgreSQL database dump complete
--

