
const {GraphQLClient, gql} = require('graphql-request');

const addNoteMutation = gql`
  mutation addNote($note: notes_insert_input! = {}) {
    insert_notes_one(object: $note) {
      id
    }
  }
`;

const addTagsAndLinks = gql`
  mutation addNoteTagsAndLinks($tags: [note_tags_insert_input!]!, $links: [note_links_insert_input!]!) {
    insert_note_tags(objects: $tags) {
      returning {
        tag
      }
    }
    insert_note_links(objects: $links) {
      returning {
        to
      }
    }
  }
`;

const tagMatch = /[#]+[A-Za-z0-9-_]+/g;
const linkMatch = /\[\[\d+\]\]/g;
exports.handler = function(context, event, callback) {
  const from = event.To;
  console.log('Message from', from);
  const to = event.From;
  console.log('Responding to', to);
  if (to !== context.allowed_number) {
    console.error('Unidentified Number', to);
    return callback(new Error('Bad sender' + to));
  }

  const gqlClient = new GraphQLClient(context.hasura_endpoint, { 
    headers: {
      'x-hasura-admin-secret': context.hasura_auth,
    }
  });
  const tags = event.Body?.match(tagMatch)?.map(tag => tag.slice(1)) || [];
  const links = event.Body?.match(linkMatch)?.map(link => link.replace(/[\[\]]/g, '')) || [];
  const variables = {
    note: {
      note: event.Body,
      creator: event.From,
    }
  };

  console.log('Note:', variables.note.note);
  console.log('Creator:', variables.note.creator);
  console.log('Tags:', tags);
  console.log('Links:', links);


  gqlClient.request(addNoteMutation, variables)
    .then((data) => {
      // add links + tags
      const id = data?.insert_notes_one?.id;
      if (!id) {
        console.log('something went wrong', data);
        return {id: 'something went wrong'};
      }

      if (!(tags.length || links.length)) {
        // there's nothing to add
        console.log('no tags/links to add')
        return {id};
      }

      const tagsAndLinks = {
        links: links.map(link => ({from: id, to: link})),
        tags: tags.map(tag => ({note_id: id, tag})),
      };

      console.log('adding tags + links to note', id)
      return gqlClient.request(addTagsAndLinks, tagsAndLinks).then(data => ({
        tags: data?.insert_note_tags?.returning?.map(({tag}) => tag),
        links: data?.insert_note_links?.returning?.map(({to}) => to),
        id,
      }));
    })
    .then(({
      id,
      tags = [],
      links = []
    }) => {
        callback(null, `Created note: ${id}
${tags.length && `tags: ${tags.join(',')}`}
${links.length && `links: ${links.join(',')}`}
    `);
      })
    .catch((error) => {
      console.error(error);
      return callback(error);
    });
};
