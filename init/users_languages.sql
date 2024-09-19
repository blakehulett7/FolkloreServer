create table users_languages(
    user_id text,
    language_id text,
    unique(user_id, language_id)
);
