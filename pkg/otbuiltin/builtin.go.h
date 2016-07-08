#include <glib.h>
#include <ostree.h>
#include <string.h>

static void
_ostree_repo_append_modifier_flags(OstreeRepoCommitModifierFlags *flags, int flag) {
  *flags |= flag;
}

struct CommitFilterData {
  GHashTable *mode_adds;
  GHashTable *skip_list;
};

typedef struct CommitFilterData CommitFilterData;

static OstreeRepoCommitFilterResult
_commit_filter (OstreeRepo         *self,
               const char          *path,
               GFileInfo           *file_info,
               gpointer            user_data,
               int                 owner_uid,
               int                 owner_gid)
{
  struct CommitFilterData *data = user_data;

  GHashTable *mode_adds = data->mode_adds;
  GHashTable *skip_list = data->skip_list;
  gpointer value;

  if (owner_uid >= 0)
    {
      g_file_info_set_attribute_uint32(file_info, "unix::uid", owner_uid);
    }
  if (owner_gid >= 0)
    {
      g_file_info_set_attribute_uint32(file_info, "unix::gid", owner_gid);
    }

  if (mode_adds && g_hash_table_lookup_extended (mode_adds, path, NULL, &value))
    {
      guint current_mode = g_file_info_get_attribute_uint32 (file_info, "unix::mode");
      guint mode_add = GPOINTER_TO_UINT (value);
      g_file_info_set_attribute_uint32 (file_info, "unix::mode",
                                        current_mode | mode_add);
      g_hash_table_remove (mode_adds, path);
    }

  if (skip_list && g_hash_table_contains (skip_list, path))
    {
      g_hash_table_remove (skip_list, path);
      return OSTREE_REPO_COMMIT_FILTER_SKIP;
    }

  return OSTREE_REPO_COMMIT_FILTER_ALLOW;
}

static char* _gptr_to_str(gpointer p)
{
    return (char*)p;
}

// The following 3 functions are wrapper functions for macros since CGO can't parse macros
static OstreeRepoFile*
_ostree_repo_file(GFile *file)
{
  return OSTREE_REPO_FILE (file);
}

static guint
_gpointer_to_uint (gpointer ptr)
{
  return GPOINTER_TO_UINT (ptr);
}

static gpointer
_guint_to_pointer (guint u)
{
  return GUINT_TO_POINTER (u);
}


// These functions are wrappers for variadic functions since CGO can't parse variadic functions
static void
_g_printerr_onearg (char* msg,
                    char* arg)
{
  g_printerr("%s %s\n", msg, arg);
}

static void
_g_set_error_onearg (GError *err,
                     char*  msg,
                     char*  arg)
{
  g_set_error(&err, G_IO_ERROR, G_IO_ERROR_FAILED, "%s %s", msg, arg);
}

static void
_g_variant_builder_add_twoargs (GVariantBuilder*     builder,
                                const gchar   *format_string,
                                char          *arg1,
                                char          *arg2)
{
  g_variant_builder_add(builder, format_string, arg1, arg2);
}
