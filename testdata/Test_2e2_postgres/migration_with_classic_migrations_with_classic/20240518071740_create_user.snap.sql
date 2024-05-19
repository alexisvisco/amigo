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
    id text NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: migrations_with_classic; Owner: -
--

CREATE TABLE migrations_with_classic.users (
    id integer NOT NULL,
    name text,
    email text,
    created_at timestamp(6) without time zone DEFAULT now() NOT NULL,
    updated_at timestamp(6) without time zone DEFAULT now() NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: migrations_with_classic; Owner: -
--

CREATE SEQUENCE migrations_with_classic.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: migrations_with_classic; Owner: -
--

ALTER SEQUENCE migrations_with_classic.users_id_seq OWNED BY migrations_with_classic.users.id;


--
-- Name: users id; Type: DEFAULT; Schema: migrations_with_classic; Owner: -
--

ALTER TABLE ONLY migrations_with_classic.users ALTER COLUMN id SET DEFAULT nextval('migrations_with_classic.users_id_seq'::regclass);


--
-- Name: mig_schema_versions mig_schema_versions_pkey; Type: CONSTRAINT; Schema: migrations_with_classic; Owner: -
--

ALTER TABLE ONLY migrations_with_classic.mig_schema_versions
    ADD CONSTRAINT mig_schema_versions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: migrations_with_classic; Owner: -
--

ALTER TABLE ONLY migrations_with_classic.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_users_name; Type: INDEX; Schema: migrations_with_classic; Owner: -
--

CREATE INDEX idx_users_name ON migrations_with_classic.users USING btree (name);


--
-- PostgreSQL database dump complete
--

