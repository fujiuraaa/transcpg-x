-- TransCPG-X · 005 · View baca, fungsi analisis klaim, dan pengamanan akses

-- Volume episode per grouper (dasar "layak disusun" vs "dilewati") -------------
create view v_grouper_volume with (security_invoker = true) as
select g.code as cbg_code, g.name, g.mdc_code, count(e.id)::int as episodes
from ina_cbg_groupers g
left join claim_episodes e on e.cbg_code = g.code
group by g.code;

-- Ringkasan satu CP untuk CP Library, Dashboard, dan Approval -----------------
create view v_pathway_overview with (security_invoker = true) as
select
  p.id, p.cbg_code, p.name, g.mdc_code, m.name as mdc_name,
  p.stage, p.stage_changed_at, p.activated_at, p.split_decision, p.has_pediatric_cohort,
  coalesce(s.episodes, 0)               as episodes,
  s.dominant_severity,
  s.target_los::float8                  as target_los,
  coalesce(s.sev_episodes, '{}'::jsonb) as sev_episodes,
  coalesce(s.severities_adequate, false) as severities_adequate,
  coalesce(r.acuan_count, 0)            as acuan_count,
  r.acuan_titles,
  coalesce(c.candidate_count, 0)        as candidate_count,
  coalesce(k.procedures, 0)             as procedure_count,
  coalesce(k.mapped, 0)                 as procedure_kptl_mapped
from clinical_pathways p
join ina_cbg_groupers g on g.code = p.cbg_code
join mdc_groups m on m.code = g.mdc_code
left join lateral (
  select sum(ss.episodes)::int                                   as episodes,
         (array_agg(ss.severity   order by ss.episodes desc))[1] as dominant_severity,
         (array_agg(ss.target_los order by ss.episodes desc))[1] as target_los,
         jsonb_object_agg(ss.severity, ss.episodes)              as sev_episodes,
         count(*) = 3 and bool_and(ss.episodes >= 10)            as severities_adequate
  from pathway_severity_stats ss
  where ss.pathway_id = p.id
) s on true
left join lateral (
  select count(*)::int as acuan_count,
         string_agg(d.title, '; ' order by pr.assigned_at) as acuan_titles
  from pathway_references pr
  join guideline_documents d on d.id = pr.document_id
  where pr.pathway_id = p.id and pr.sub_cp_id is null
) r on true
left join lateral (
  select count(*)::int as candidate_count
  from guideline_documents d
  where d.icd10_codes && array(
    select distinct cd.icd10_code
    from claim_episodes e
    join claim_episode_diagnoses cd on cd.episode_id = e.id and cd.is_primary
    where e.cbg_code = p.cbg_code)
) c on true
left join lateral (
  select count(distinct cp.icd9cm_code)::int as procedures,
         count(distinct km.icd9cm_code)::int as mapped
  from claim_episodes e
  join claim_episode_procedures cp on cp.episode_id = e.id
  left join kptl_mappings km on km.icd9cm_code = cp.icd9cm_code
  where e.cbg_code = p.cbg_code
) k on true;

-- Langkah 1 alur utama: grouper dengan episode ≥5 menjadi CP Draf ---------------
create or replace function generate_draft_pathways(min_episodes int default 5) returns integer
language sql as $$
  with inserted as (
    insert into clinical_pathways (cbg_code, name)
    select v.cbg_code, v.name
    from v_grouper_volume v
    where v.episodes >= min_episodes
    on conflict (cbg_code) do nothing
    returning 1
  )
  select count(*)::int from inserted;
$$;

-- Statistik baseline per severity dari klaim SEBELUM CP aktif -------------------
-- Catatan: target_los sementara = median LOS. Aturan final ditetapkan Tim CP.
create or replace function refresh_pathway_severity_stats() returns void
language sql as $$
  insert into pathway_severity_stats (pathway_id, severity, episodes, los_median, los_p75, target_los, age_median)
  select p.id, e.severity, count(*),
         percentile_cont(0.5)  within group (order by e.los_days),
         percentile_cont(0.75) within group (order by e.los_days),
         percentile_cont(0.5)  within group (order by e.los_days),
         percentile_cont(0.5)  within group (order by e.age_years)
  from clinical_pathways p
  join claim_episodes e on e.cbg_code = p.cbg_code
  where p.activated_at is null or e.discharge_date < p.activated_at::date
  group by p.id, e.severity
  on conflict (pathway_id, severity) do update set
    episodes   = excluded.episodes,
    los_median = excluded.los_median,
    los_p75    = excluded.los_p75,
    target_los = excluded.target_los,
    age_median = excluded.age_median;
$$;

-- Pengamanan: seluruh akses data lewat backend Go (koneksi pemilik tabel). ---------
-- RLS aktif tanpa policy → API REST/anon Supabase tidak bisa membaca apa pun.
do $$
declare t record;
begin
  for t in select tablename from pg_tables where schemaname = 'public' loop
    execute format('alter table public.%I enable row level security', t.tablename);
  end loop;
  if exists (select 1 from pg_roles where rolname = 'anon') then
    execute 'revoke all on all tables in schema public from anon, authenticated';
    execute 'revoke all on all functions in schema public from anon, authenticated';
  end if;
end $$;
