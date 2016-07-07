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
               const char         *path,
               GFileInfo          *file_info,
               gpointer            user_data)
{
  struct CommitFilterData *data = user_data;
  return commitFilter (self, path, file_info, data);
}

static char* _gptr_to_str(gpointer p)
{
    return (char*)p;
}

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
