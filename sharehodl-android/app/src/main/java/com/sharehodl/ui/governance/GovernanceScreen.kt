package com.sharehodl.ui.governance

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.sharehodl.service.Proposal
import com.sharehodl.viewmodel.WalletViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GovernanceScreen(
    viewModel: WalletViewModel
) {
    val proposals by viewModel.proposals.collectAsState()
    val uiState by viewModel.uiState.collectAsState()

    var selectedProposal by remember { mutableStateOf<Proposal?>(null) }
    var showVoteDialog by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        viewModel.fetchProposals()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Governance") },
                actions = {
                    IconButton(onClick = { viewModel.fetchProposals() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        }
    ) { padding ->
        if (proposals.isEmpty() && !uiState.isLoading) {
            EmptyProposals(modifier = Modifier.padding(padding))
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                // Stats Card
                item {
                    GovernanceStatsCard(totalProposals = proposals.size)
                }

                // Active Proposals Header
                item {
                    Text(
                        text = "Proposals",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold
                    )
                }

                items(proposals) { proposal ->
                    ProposalCard(
                        proposal = proposal,
                        onVote = {
                            selectedProposal = proposal
                            showVoteDialog = true
                        }
                    )
                }
            }
        }

        // Loading
        if (uiState.isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator()
            }
        }
    }

    // Vote Dialog
    if (showVoteDialog && selectedProposal != null) {
        VoteDialog(
            proposal = selectedProposal!!,
            onDismiss = {
                showVoteDialog = false
                selectedProposal = null
            },
            onVote = { vote ->
                // viewModel.vote(activity, selectedProposal!!.id, vote)
                showVoteDialog = false
                selectedProposal = null
            }
        )
    }
}

@Composable
fun GovernanceStatsCard(totalProposals: Int) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.primary
        )
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp),
            horizontalArrangement = Arrangement.SpaceAround
        ) {
            StatItem(
                label = "Total Proposals",
                value = totalProposals.toString()
            )
            StatItem(
                label = "Quorum Required",
                value = "33.4%"
            )
            StatItem(
                label = "Pass Threshold",
                value = "50%"
            )
        }
    }
}

@Composable
fun StatItem(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = value,
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.onPrimary
        )
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onPrimary.copy(alpha = 0.8f)
        )
    }
}

@Composable
fun ProposalCard(
    proposal: Proposal,
    onVote: () -> Unit
) {
    val isVoting = proposal.status == "PROPOSAL_STATUS_VOTING_PERIOD"

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Proposal ID
                Text(
                    text = "#${proposal.id}",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.primary
                )

                // Status Badge
                StatusBadge(status = proposal.displayStatus)
            }

            Spacer(modifier = Modifier.height(8.dp))

            // Title
            Text(
                text = proposal.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(4.dp))

            // Description
            if (proposal.description.isNotEmpty()) {
                Text(
                    text = proposal.description.take(150) + if (proposal.description.length > 150) "..." else "",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            // Vote Button (only for active proposals)
            if (isVoting) {
                Button(
                    onClick = onVote,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Icon(Icons.Default.HowToVote, contentDescription = null)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("Vote")
                }
            }
        }
    }
}

@Composable
fun StatusBadge(status: String) {
    val (backgroundColor, textColor) = when (status) {
        "Voting" -> Color(0xFF3B82F6) to Color.White
        "Passed" -> Color(0xFF22C55E) to Color.White
        "Rejected" -> Color(0xFFEF4444) to Color.White
        "Deposit" -> Color(0xFFF59E0B) to Color.Black
        else -> MaterialTheme.colorScheme.outline to MaterialTheme.colorScheme.onSurface
    }

    Surface(
        color = backgroundColor,
        shape = RoundedCornerShape(4.dp)
    ) {
        Text(
            text = status,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
            style = MaterialTheme.typography.labelSmall,
            color = textColor
        )
    }
}

@Composable
fun EmptyProposals(modifier: Modifier = Modifier) {
    Box(
        modifier = modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                Icons.Default.HowToVote,
                contentDescription = null,
                modifier = Modifier.size(64.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "No proposals yet",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                text = "Governance proposals will appear here",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f)
            )
        }
    }
}

enum class VoteOption {
    Yes, No, NoWithVeto, Abstain
}

@Composable
fun VoteDialog(
    proposal: Proposal,
    onDismiss: () -> Unit,
    onVote: (VoteOption) -> Unit
) {
    var selectedVote by remember { mutableStateOf<VoteOption?>(null) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Vote on Proposal #${proposal.id}") },
        text = {
            Column {
                Text(
                    text = proposal.title,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.Medium
                )

                Spacer(modifier = Modifier.height(16.dp))

                Text(
                    text = "Select your vote:",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Spacer(modifier = Modifier.height(8.dp))

                VoteOption.entries.forEach { option ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 4.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        RadioButton(
                            selected = selectedVote == option,
                            onClick = { selectedVote = option }
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = when (option) {
                                VoteOption.Yes -> "Yes"
                                VoteOption.No -> "No"
                                VoteOption.NoWithVeto -> "No With Veto"
                                VoteOption.Abstain -> "Abstain"
                            },
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = { selectedVote?.let { onVote(it) } },
                enabled = selectedVote != null
            ) {
                Text("Submit Vote")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}
