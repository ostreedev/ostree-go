#include <ostree.h>
#include <string.h>

static void
_ostree_repo_append_modifier_flags(OstreeRepoCommitModifierFlags *flags, int flag) {
  *flags |= flag;
}

static gboolean
_handle_statoverride_line (const char  *line,
                          void        *data,
                          GError     **error)
{
  GHashTable *files = data;
  const char *spc;
  guint mode_add;

  spc = strchr (line, ' ');
  if (spc == NULL)
    {
      g_set_error (error, G_IO_ERROR, G_IO_ERROR_FAILED,
                   "Malformed statoverride file (no space found)");
      return FALSE;
    }

  mode_add = (guint32)(gint32)g_ascii_strtod (line, NULL);
  g_hash_table_insert (files, g_strdup (spc + 1),
                       GUINT_TO_POINTER((gint32)mode_add));
  return TRUE;
}

static gboolean
_handle_skiplist_line (const char  *line,
                      void        *data,
                      GError     **error)
{
  GHashTable *files = data;
  g_hash_table_add (files, g_strdup (line));
  return TRUE;
}
