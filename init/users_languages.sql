create table users_languages(
    user_id text,
    language_id text,
    best_listening_streak integer,
    current_listening_streak integer,
    words_learned integer,
    last_listened_at text,
    unique(user_id, language_id)
);
