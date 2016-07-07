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

static OstreeRepoCommitFilterResult
_commit_filter (OstreeRepo         *self,
               const char         *path,
               GFileInfo          *file_info,
               gpointer            user_data)
{
  struct CommitFilterData *data = user_data;
  return commitFilter (self, path, file_info, data);
  /*GHashTable *mode_adds = data->mode_adds;
  GHashTable *skip_list = data->skip_list;
  gpointer value;

  if (opt_owner_uid >= 0)
    g_file_info_set_attribute_uint32 (file_info, "unix::uid", opt_owner_uid);
  if (opt_owner_gid >= 0)
    g_file_info_set_attribute_uint32 (file_info, "unix::gid", opt_owner_gid);

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
  */
}

static char* _gptr_to_str(gpointer p)
  {
    return (char*)p;
  }
