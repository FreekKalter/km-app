--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: docker; Tablespace: 
--

CREATE TABLE goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now()
);


ALTER TABLE public.goose_db_version OWNER TO docker;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: docker
--

CREATE SEQUENCE goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.goose_db_version_id_seq OWNER TO docker;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: docker
--

ALTER SEQUENCE goose_db_version_id_seq OWNED BY goose_db_version.id;


--
-- Name: kilometers; Type: TABLE; Schema: public; Owner: docker; Tablespace: 
--

CREATE TABLE kilometers (
    id integer NOT NULL,
    date date,
    begin integer,
    eerste integer,
    laatste integer,
    terug integer,
    comment character varying(200)
);


ALTER TABLE public.kilometers OWNER TO docker;

--
-- Name: kilometers_id_seq; Type: SEQUENCE; Schema: public; Owner: docker
--

CREATE SEQUENCE kilometers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.kilometers_id_seq OWNER TO docker;

--
-- Name: kilometers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: docker
--

ALTER SEQUENCE kilometers_id_seq OWNED BY kilometers.id;


--
-- Name: times; Type: TABLE; Schema: public; Owner: docker; Tablespace: 
--

CREATE TABLE times (
    id integer NOT NULL,
    date date,
    checkin integer,
    checkout integer
);


ALTER TABLE public.times OWNER TO docker;

--
-- Name: times_id_seq; Type: SEQUENCE; Schema: public; Owner: docker
--

CREATE SEQUENCE times_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.times_id_seq OWNER TO docker;

--
-- Name: times_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: docker
--

ALTER SEQUENCE times_id_seq OWNED BY times.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: docker
--

ALTER TABLE ONLY goose_db_version ALTER COLUMN id SET DEFAULT nextval('goose_db_version_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: docker
--

ALTER TABLE ONLY kilometers ALTER COLUMN id SET DEFAULT nextval('kilometers_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: docker
--

ALTER TABLE ONLY times ALTER COLUMN id SET DEFAULT nextval('times_id_seq'::regclass);


--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: docker
--

COPY goose_db_version (id, version_id, is_applied, tstamp) FROM stdin;
1	0	t	2013-12-11 20:45:24.857764
2	20131211211952	t	2013-12-11 20:45:24.872793
3	20131211214013	t	2013-12-11 20:45:24.888451
4	20131211214650	t	2013-12-11 20:50:12.658567
5	20131211214650	f	2013-12-11 20:50:31.920659
6	20131211214013	f	2013-12-11 20:50:34.416322
7	20131211211952	f	2013-12-11 20:50:35.176186
8	20131211211952	t	2013-12-11 20:50:39.48731
9	20131211214013	t	2013-12-11 20:50:39.503629
10	20131211214650	t	2013-12-11 20:50:39.519191
37	20131211214650	f	2013-12-11 20:53:38.221065
38	20131211214013	f	2013-12-11 20:53:40.315182
39	20131211211952	f	2013-12-11 20:53:41.080449
40	20131211211952	t	2013-12-11 20:53:46.185741
41	20131211214013	t	2013-12-11 20:53:46.202562
42	20131211214650	t	2013-12-11 20:53:46.216497
70	20131211214650	f	2013-12-11 20:55:39.800097
71	20131211214013	f	2013-12-11 20:55:40.997172
72	20131211211952	f	2013-12-11 20:55:41.330007
73	20131211211952	t	2013-12-11 20:55:44.381361
74	20131211214013	t	2013-12-11 20:55:44.399543
75	20131211214650	t	2013-12-11 20:55:44.412439
76	20131211214650	f	2013-12-11 21:56:16.763793
77	20131211214013	f	2013-12-11 21:56:17.510949
78	20131211211952	f	2013-12-11 21:56:18.170094
79	20131211211952	t	2013-12-11 21:58:15.136355
80	20131211214013	t	2013-12-11 21:58:15.158571
109	20131211214013	f	2013-12-11 22:18:48.611331
110	20131211211952	f	2013-12-11 22:18:49.700075
111	20131211211952	t	2013-12-11 22:18:53.819918
112	20131211214013	t	2013-12-11 22:18:53.845799
142	20131211214013	f	2013-12-11 22:19:59.246549
143	20131211211952	f	2013-12-11 22:20:05.892217
144	20131211211952	t	2013-12-11 22:20:10.604158
145	20131211214013	t	2013-12-11 22:20:10.615349
146	20131211214013	f	2013-12-11 22:22:28.247051
147	20131211211952	f	2013-12-11 22:22:29.210418
175	20131211211952	t	2013-12-11 22:22:37.178736
176	20131211214013	t	2013-12-11 22:22:37.200924
177	20131211214013	f	2013-12-12 21:44:10.230573
178	20131211211952	f	2013-12-12 21:44:11.234334
179	20131211211952	t	2013-12-12 21:44:14.780885
180	20131211214013	t	2013-12-12 21:44:14.818021
181	20131211214013	f	2013-12-12 22:16:56.739965
182	20131211211952	f	2013-12-12 22:16:58.744212
183	20131211211952	t	2013-12-12 22:17:00.656494
184	20131211214013	t	2013-12-12 22:17:14.122802
185	20131211214013	f	2013-12-12 22:18:20.832801
186	20131211211952	f	2013-12-12 22:18:21.360281
187	20131211211952	t	2013-12-12 22:18:23.567175
188	20131211214013	t	2013-12-12 22:18:23.584997
189	20131211214013	f	2013-12-13 19:49:11.09969
190	20131211211952	f	2013-12-13 19:49:12.947259
191	20131211211952	t	2013-12-13 19:49:31.182343
192	20131211214013	t	2013-12-13 19:49:31.216447
193	20131211214013	f	2013-12-24 20:43:29.25651
194	20131211211952	f	2013-12-24 20:43:34.214829
195	20131211211952	t	2013-12-24 20:43:38.934474
196	20131211214013	t	2013-12-24 20:43:38.954939
197	20131211214013	f	2013-12-24 20:50:00.231298
198	20131211211952	f	2013-12-24 20:50:01.276051
199	20131211211952	t	2013-12-24 20:50:05.055455
200	20131211214013	t	2013-12-24 20:50:05.073835
\.


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: docker
--

SELECT pg_catalog.setval('goose_db_version_id_seq', 200, true);


--
-- Data for Name: kilometers; Type: TABLE DATA; Schema: public; Owner: docker
--

COPY kilometers (id, date, begin, eerste, laatste, terug, comment) FROM stdin;
2	2013-11-02	14093	14109	14186	14188	
3	2013-11-03	14188	14195	14281	14296	
4	2013-11-04	14296	14306	14395	14405	
5	2013-11-05	14405	14428	14521	14530	
6	2013-11-06	14530	14555	14597	14606	
7	2013-11-07	14606	14617	14666	14683	
8	2013-11-08	14683	14731	14756	14761	
9	2013-11-09	14761	14761	14831	14832	
10	2013-11-10	14832	14835	14847	14853	
11	2013-11-11	14853	14885	14913	14928	
12	2013-11-12	14928	14946	14956	14970	
13	2013-11-13	14970	14976	15038	15056	
14	2013-11-14	15056	15102	15184	15195	
15	2013-11-15	15195	15243	15335	15340	
16	2013-11-16	15340	15375	15429	15434	
17	2013-11-17	15434	15448	15464	15482	
18	2013-11-18	15482	15522	15531	15546	
19	2013-11-19	15546	15568	15643	15661	
20	2013-11-20	15661	15699	15699	15702	
21	2013-11-21	15702	15746	15813	15826	
22	2013-11-22	15826	15868	15887	15901	
23	2013-11-23	15901	15923	15949	15966	
24	2013-11-24	15966	16007	16058	16061	
25	2013-11-25	16061	16109	16165	16172	
26	2013-11-26	16172	16209	16224	16224	
27	2013-11-27	16224	16243	16295	16314	
28	2013-11-28	16314	16319	16380	16394	
29	2013-11-29	16394	16437	16533	16545	
91	2013-12-25	16545	163233	0	0	
\.


--
-- Name: kilometers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: docker
--

SELECT pg_catalog.setval('kilometers_id_seq', 91, true);


--
-- Data for Name: times; Type: TABLE DATA; Schema: public; Owner: docker
--

COPY times (id, date, checkin, checkout) FROM stdin;
1	2013-12-25	1388011124	0
\.


--
-- Name: times_id_seq; Type: SEQUENCE SET; Schema: public; Owner: docker
--

SELECT pg_catalog.setval('times_id_seq', 1, true);


--
-- Name: goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: docker; Tablespace: 
--

ALTER TABLE ONLY goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: pm; Type: CONSTRAINT; Schema: public; Owner: docker; Tablespace: 
--

ALTER TABLE ONLY kilometers
    ADD CONSTRAINT pm PRIMARY KEY (id);


--
-- Name: times_pkey; Type: CONSTRAINT; Schema: public; Owner: docker; Tablespace: 
--

ALTER TABLE ONLY times
    ADD CONSTRAINT times_pkey PRIMARY KEY (id);


--
-- Name: unique_date; Type: CONSTRAINT; Schema: public; Owner: docker; Tablespace: 
--

ALTER TABLE ONLY kilometers
    ADD CONSTRAINT unique_date UNIQUE (date);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

